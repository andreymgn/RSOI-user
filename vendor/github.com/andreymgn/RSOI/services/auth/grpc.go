package auth

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

const DefaultExpirationTime = time.Minute * 15

var (
	ErrNotFound    = errors.New("no app with this ID")
	ErrWrongSecret = errors.New("secret doesn't match")
)

type InternalAPITokenStorage struct {
	redis     *redis.Client
	knownApps map[string]string
}

func generateToken() string {
	return uuid.New().String()
}

// NewTokenStorage return new instance of token storage
func NewInternalAPITokenStorage(addr, password string, db int, knownApps map[string]string) (*InternalAPITokenStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping().Result()
	return &InternalAPITokenStorage{client, knownApps}, err
}

func (s *InternalAPITokenStorage) Add(appID, appSecret string) (string, error) {
	secret, ok := s.knownApps[appID]
	if !ok {
		return "", ErrNotFound
	}

	if secret != appSecret {
		return "", ErrWrongSecret
	}

	t := generateToken()
	err := s.redis.Set(t, true, DefaultExpirationTime).Err()
	return t, err
}

func (s *InternalAPITokenStorage) Exists(token string) (bool, error) {
	exists, err := s.redis.Exists(token).Result()
	return exists == 1, err
}
