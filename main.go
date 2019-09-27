package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/hugoatease/garfunkel/clients"
	"github.com/hugoatease/garfunkel/credentials"
	"github.com/hugoatease/garfunkel/queue"
	kafka "github.com/segmentio/kafka-go"
)

var (
	spotifyClient   clients.Client = clients.NewSpotifyClient(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))
	deezerClient    clients.Client = clients.NewDeezerClient()
	clientsServices                = map[queue.Service]clients.Client{
		queue.Spotify: spotifyClient,
		queue.Deezer:  deezerClient,
	}
)

func getCurrentlyPlaying(item queue.QueueItem, conn redis.Conn) (*clients.Listen, error) {
	credentialsStores := map[queue.Service]credentials.CredentialsStore{
		queue.Spotify: credentials.NewSpotifyStore(conn),
		queue.Deezer:  credentials.NewDeezerStore(conn),
	}

	client := clientsServices[item.Service]
	store := credentialsStores[item.Service]

	var listen *clients.Listen

	creds, err := store.Get(item.UserId)
	if err != nil {
		return nil, err
	}

	listen, err = client.GetCurrentlyPlaying(creds.Token)
	if err != nil {
		if err != clients.ExpiredToken {
			return nil, err
		}

		tokenResponse, err := client.(*clients.Spotify).RefreshAccessToken(creds.RefreshToken)
		if err != nil {
			fmt.Printf("%s", err)
			return nil, err
		}
		creds.Token = tokenResponse.AccessToken
		creds.ExpiresAt = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
		store.(*credentials.SpotifyStore).Set(creds)

		return getCurrentlyPlaying(item, conn)
	}

	return listen, nil
}

func main() {
	redisPool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(os.Getenv("REDIS_URL")) },
	}

	conn := redisPool.Get()
	defer conn.Close()
	queueConn := redisPool.Get()
	defer queueConn.Close()

	q := queue.NewQueue(queueConn, 500*time.Millisecond)
	ch := make(chan queue.QueueItem)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "garfunkel",
		Balancer: &kafka.LeastBytes{},
	})

	defer w.Close()

	go q.Poll(ch)

	for item := range ch {
		listen, err := getCurrentlyPlaying(item, conn)
		if err != nil {
			continue
		}

		value, err := json.Marshal(listen)
		if err != nil {
			fmt.Printf("%+v", err)
			continue
		}

		w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(item.UserId),
				Value: value,
			},
		)

		fmt.Printf("%+v", listen)
	}
}
