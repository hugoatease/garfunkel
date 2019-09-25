package clients

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Deezer struct {
	Client       *http.Client
	ClientID     string
	ClientSecret string
}

func NewDeezerClient() *Deezer {
	deezer := &Deezer{
		Client: &http.Client{},
	}
	return deezer
}

func (c *Deezer) GetCurrentlyPlaying(token string) (*DeezerTrack, error) {
	res, err := http.Get("https://api.deezer.com/user/me/history?access_token=" + token)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var history DeezerHistory
	err = json.Unmarshal(body, &history)
	if err != nil {
		return nil, err
	}

	return &history.Data[0], nil
}
