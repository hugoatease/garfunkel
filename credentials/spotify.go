package credentials

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// SpotifyStore fetches Spotify user credentials from Redis
type SpotifyStore struct {
	Client *redis.Client
}

// NewSpotifyStore creates a Spotify store
func NewSpotifyStore(options redis.Options) *SpotifyStore {
	return &SpotifyStore{
		Client: redis.NewClient(&options),
	}
}

func (s *SpotifyStore) Get(userId string) (*SpotifyCredentials, error) {
	tokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".token"}, "")
	refreshTokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".refresh-token"}, "")
	expiresAtKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".expires"}, "")

	token, err := s.Client.Get(tokenKey).Result()
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.Client.Get(refreshTokenKey).Result()
	if err != nil {
		return nil, err
	}

	expiresAtTimestamp, err := s.Client.Get(expiresAtKey).Result()
	if err != nil {
		return nil, err
	}

	expiresAt, err := strconv.ParseInt(expiresAtTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}

	return &SpotifyCredentials{
		UserId:       userId,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
	}, nil
}

func (s *SpotifyStore) Set(credentials *SpotifyCredentials) error {
	tokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".token"}, "")
	refreshTokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".refresh-token"}, "")
	expiresAtKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".expires"}, "")

	expiresAt := strconv.FormatInt(credentials.ExpiresAt.Unix(), 10)

	pipe := s.Client.TxPipeline()
	pipe.Set(tokenKey, credentials.Token, 0)
	pipe.Set(refreshTokenKey, credentials.RefreshToken, 0)
	pipe.Set(expiresAtKey, expiresAt, 0)
	_, err := pipe.Exec()

	return err
}
