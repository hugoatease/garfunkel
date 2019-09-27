package credentials

import "time"

type CredentialsStore interface {
	Get(userId string) (*Credentials, error)
}

// SpotifyCredentials contains tokens used to make requests to the Spotify API on a user's behalf
type SpotifyCredentials struct {
	UserId       string
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

type DeezerCredentials struct {
	Token string
}

type Credentials struct {
	UserId       string
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}
