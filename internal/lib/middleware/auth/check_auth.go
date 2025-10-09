package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type SessionData struct {
	UserID int `json:"user_id"`
}

type AuthenticatorInterface interface {
	Authenticate(cookie string) (userID int, err error)
}

type Authenticator struct {
	redisClient *redis.Client
}

func NewAuthenticator(redisClient *redis.Client) *Authenticator {
	return &Authenticator{redisClient: redisClient}
}

func (a *Authenticator) Authenticate(cookie string) (int, error) {
	ctx := context.Background()

	val, err := a.redisClient.Get(ctx, cookie).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrSessionNotFound
		}
		return 0, fmt.Errorf("%w: %v", ErrRedisRead, err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrJSONDecode, err)
	}

	return session.UserID, nil
}
