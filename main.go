package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/hugoatease/garfunkel/clients"
	"github.com/hugoatease/garfunkel/credentials"
	"github.com/hugoatease/garfunkel/queue"
)

func main() {
	conn, _ := redis.DialURL("redis://")
	conn2, _ := redis.DialURL("redis://")
	q := queue.NewQueue(conn, 500*time.Millisecond)
	creds := credentials.NewSpotifyStore(conn2)
	ch := make(chan queue.QueueItem)

	client := clients.NewClient(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))

	go q.Poll(ch)

	for item := range ch {
		spotifyCredentials, err := creds.Get(item.UserId)
		var listen *clients.SpotifyListen
		if err == nil {
			listen, err = client.GetCurrentlyPlaying(spotifyCredentials.Token)
			if err != nil && err == clients.ExpiredToken {
				tokenResponse, err := client.RefreshAccessToken(spotifyCredentials.RefreshToken)
				if err != nil {
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
		}
		fmt.Printf("%+v", listen)
	}
}
