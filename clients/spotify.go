package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Spotify struct {
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
	Client       *http.Client
}

func NewClient(token string, refreshToken string, expiresAt time.Time) *Spotify {
	spotify := &Spotify{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Client:       &http.Client{},
	}
	return spotify
}

func (c *Spotify) GetRecentlyPlayed() (*SpotifyListen, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/recently-played", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	listen := new(SpotifyListen)
	json.Unmarshal(body, listen)

	return listen, nil
}
