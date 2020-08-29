package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/hugoatease/garfunkel/clients"
	"github.com/hugoatease/garfunkel/credentials"
	"github.com/hugoatease/garfunkel/queue"
	kafka "github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
)

type clientInstances map[queue.Service]clients.Client
type credentialsInstances map[queue.Service]credentials.CredentialsStore

func getCurrentlyPlaying(item queue.QueueItem, clientsServices clientInstances, credentialsStores credentialsInstances) (*clients.Listen, error) {
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

		return getCurrentlyPlaying(item, clientsServices, credentialsStores)
	}

	return listen, nil
}

func createMqttClient(url string) (*mqtt.Client, mqtt.Token) {
	mqttClientOpts := mqtt.NewClientOptions()
	mqttClientOpts.AddBroker(url)
	mqttClientOpts.SetAutoReconnect(true)
	mqttClient := mqtt.NewClient(mqttClientOpts)
	connectToken := mqttClient.Connect()
	connectToken.Wait()
	return &mqttClient, connectToken
}

func fetchListens(c *cli.Context) error {
	var (
		kafkaWriter *kafka.Writer
		mqttClient  *mqtt.Client
	)

	redisPool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(c.String("redis-url")) },
	}

	conn := redisPool.Get()
	defer conn.Close()
	queueConn := redisPool.Get()
	defer queueConn.Close()

	credentialsStores := credentialsInstances{
		queue.Spotify: credentials.NewSpotifyStore(conn),
		queue.Deezer:  credentials.NewDeezerStore(conn),
	}

	clientsServices := clientInstances{
		queue.Deezer: clients.NewDeezerClient(),
	}

	if c.IsSet("spotify-api-id") && c.IsSet("spotify-api-secret") {
		spotifyClient := clients.NewSpotifyClient(c.String("spotify-api-id"), c.String("spotify-api-secret"))
		clientsServices[queue.Spotify] = spotifyClient
	}

	if c.IsSet("kafka-url") {
		kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{c.String("kafka-url")},
			Topic:    "garfunkel",
			Balancer: &kafka.LeastBytes{},
		})

		defer kafkaWriter.Close()
	}

	if c.IsSet("mqtt-url") {
		var connectToken mqtt.Token
		mqttClient, connectToken = createMqttClient(c.String("mqtt-url"))
		err := connectToken.Error()
		if err != nil {
			fmt.Printf("%+v", err)
			return err
		}
	}

	if !c.IsSet("kafka-url") && !c.IsSet("mqtt-url") {
		fmt.Print("Error: Kafka or MQTT broker URL must be specified\n\n")
		cli.ShowAppHelpAndExit(c, 1)
	}

	q := queue.NewQueue(queueConn, 500*time.Millisecond)
	ch := make(chan queue.QueueItem)

	go q.Poll(ch)

	for item := range ch {
		listen, err := getCurrentlyPlaying(item, clientsServices, credentialsStores)
		if err != nil {
			continue
		}

		value, err := json.Marshal(listen)
		if err != nil {
			fmt.Printf("%+v", err)
			continue
		}

		if kafkaWriter != nil {
			kafkaWriter.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(item.UserId),
					Value: value,
				},
			)
		}

		if mqttClient != nil {
			(*mqttClient).Publish("garfunkel/status/"+item.UserId, 0, false, value)
		}

		fmt.Printf("%+v\n", listen)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "garfunkel",
		Usage: "publish Spotify/Deezer listens to Kafka/MQTT",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "spotify-api-id",
				EnvVars: []string{"SPOTIFY_API_ID"},
			},
			&cli.StringFlag{
				Name:    "spotify-api-secret",
				EnvVars: []string{"SPOTIFY_API_SECRET"},
			},
			&cli.StringFlag{
				Name:    "redis-url",
				EnvVars: []string{"REDIS_URL"},
				Value:   "redis://localhost:6379",
			},
			&cli.StringFlag{
				Name:    "kafka-url",
				EnvVars: []string{"KAFKA_URL"},
			},
			&cli.StringFlag{
				Name:    "mqtt-url",
				EnvVars: []string{"MQTT_URL"},
			},
		},
		Action: fetchListens,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
