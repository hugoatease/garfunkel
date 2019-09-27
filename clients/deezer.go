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

func convertDeezerHistory(item DeezerTrack) Listen {
	return Listen{
		ArtistName: item.Artist.Name,
		AlbumName:  item.Album.Title,
		TrackName:  item.Title,
		Timestamp:  item.Timestamp,
	}
}

func (c *Deezer) GetCurrentlyPlaying(token string) (*Listen, error) {
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

	item := convertDeezerHistory(history.Data[0])
	return &item, nil
}
