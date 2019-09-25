package credentials

import (
	"strings"

	"github.com/gomodule/redigo/redis"
)

type DeezerStore struct {
	Client redis.Conn
}

func NewDeezerStore(conn redis.Conn) *DeezerStore {
	return &DeezerStore{
		Client: conn,
	}
}

func (s *DeezerStore) Get(userId string) (*DeezerCredentials, error) {
	tokenKey := strings.Join([]string{"garfunkel.credentials.deezer-", userId, ".token"}, "")

	token, err := redis.String(s.Client.Do("GET", tokenKey))
	if err != nil {
		return nil, err
	}

	return &DeezerCredentials{
		Token: token,
	}, nil
}
