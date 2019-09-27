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
	creds := credentials.NewSpotifyStore(conn)
	deezerCreds := credentials.NewDeezerStore(conn)
	ch := make(chan queue.QueueItem)

	client := clients.NewSpotifyClient(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))
	deezerClient := clients.NewDeezerClient()

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "garfunkel",
		Balancer: &kafka.LeastBytes{},
	})

	defer w.Close()

	go q.Poll(ch)

	for item := range ch {
		var listen interface{}
		switch service := item.Service; service {
		case "spotify":
			spotifyCredentials, err := creds.Get(item.UserId)
			if err == nil {
				listen, err = client.GetCurrentlyPlaying(spotifyCredentials.Token)
				if err != nil && err == clients.ExpiredToken {
					fmt.Printf("%s", err)
					tokenResponse, err := client.RefreshAccessToken(spotifyCredentials.RefreshToken)
					if err != nil {
						fmt.Printf("%s", err)
						continue
					}
					spotifyCredentials.Token = tokenResponse.AccessToken
					spotifyCredentials.ExpiresAt = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
					creds.Set(spotifyCredentials)
					listen, err = client.GetCurrentlyPlaying(tokenResponse.AccessToken)
					if err != nil {
						continue
					}
				}
			} else {
				fmt.Printf("%s", err)
			}

		case "deezer":
			deezerCredentials, err := deezerCreds.Get(item.UserId)
			if err == nil {
				listen, err = deezerClient.GetCurrentlyPlaying(deezerCredentials.Token)
				if err != nil {
					fmt.Printf("%s", err)
					continue
				}
			} else {
				fmt.Printf("%s", err)
			}
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
