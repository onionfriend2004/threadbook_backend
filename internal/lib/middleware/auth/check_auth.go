package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type SessionData struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

type AuthenticatorInterface interface {
	Authenticate(cookie string) (userID uint, username string, err error)
}

type Authenticator struct {
	redisClient *redis.Client
}

func NewAuthenticator(redisClient *redis.Client) *Authenticator {
	return &Authenticator{redisClient: redisClient}
}

func (a *Authenticator) Authenticate(cookie string) (uint, string, error) {
	ctx := context.Background()
	key := "session_id:" + cookie
	val, err := a.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, "", ErrSessionNotFound
		}
		return 0, "", fmt.Errorf("%w: %v", ErrRedisRead, err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrJSONDecode, err)
	}
	fmt.Println(session)
	return session.UserID, session.Username, nil
}
