package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Spotify struct {
	Client       *http.Client
	ClientId     string
	ClientSecret string
}

var (
	ExpiredToken = errors.New("The access token expired")
)

func NewClient(clientId string, clientSecret string) *Spotify {
	spotify := &Spotify{
		Client:       &http.Client{},
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
	return spotify
}

func (c *Spotify) GetCurrentlyPlaying(token string) (*SpotifyListen, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 {
		return nil, ExpiredToken
	}

	listen := new(SpotifyListen)
	err = json.Unmarshal(body, listen)
	if err != nil {
		return nil, err
	}

	return listen, nil
}

func (c *Spotify) RefreshAccessToken(refreshToken string) (*SpotifyRefreshTokenResponse, error) {
	requestBody := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(requestBody.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.ClientId, c.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	tokenResponse := new(SpotifyRefreshTokenResponse)
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}
