package credentials

import (
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// SpotifyStore fetches Spotify user credentials from Redis
type SpotifyStore struct {
	Client redis.Conn
}

// NewSpotifyStore creates a Spotify store
func NewSpotifyStore(conn redis.Conn) *SpotifyStore {
	return &SpotifyStore{
		Client: conn,
	}
}

func (s *SpotifyStore) Get(userId string) (*Credentials, error) {
	tokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".token"}, "")
	refreshTokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".refresh-token"}, "")
	//expiresAtKey := strings.Join([]string{"garfunkel.credentials.spotify-", userId, ".expires"}, "")

	token, err := redis.String(s.Client.Do("GET", tokenKey))
	if err != nil {
		return nil, err
	}

	refreshToken, err := redis.String(s.Client.Do("GET", refreshTokenKey))
	if err != nil {
		return nil, err
	}

	/*expiresAtTimestamp, err := redis.String(s.Client.Do("GET", expiresAtKey))
	if err != nil {
		return nil, err
	}

	expiresAt, err := strconv.ParseInt(expiresAtTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}*/

	return &Credentials{
		UserId:       userId,
		Token:        token,
		RefreshToken: refreshToken,
		//ExpiresAt:    time.Unix(expiresAt, 0),
	}, nil
}

func (s *SpotifyStore) Set(credentials *Credentials) error {
	tokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".token"}, "")
	refreshTokenKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".refresh-token"}, "")
	expiresAtKey := strings.Join([]string{"garfunkel.credentials.spotify-", credentials.UserId, ".expires"}, "")

	expiresAt := strconv.FormatInt(credentials.ExpiresAt.Unix(), 10)

	s.Client.Send("SET", tokenKey, credentials.Token)
	s.Client.Send("SET", refreshTokenKey, credentials.RefreshToken)
	s.Client.Send("SET", expiresAtKey, expiresAt)
	err := s.Client.Flush()

	return err
}
