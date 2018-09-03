package clients

import "time"

type SpotifyArtist struct {
	Name string
}

type SpotifyTrack struct {
	Name    string
	Artists []SpotifyArtist
}

type SpotifyListen struct {
	PlayedAt time.Time
	Track    SpotifyTrack
}
