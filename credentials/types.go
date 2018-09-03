package credentials

import "time"

// SpotifyCredentials contains tokens used to make requests to the Spotify API on a user's behalf
type SpotifyCredentials struct {
	UserId       string
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}
