package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type SessionData struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

type AuthenticatorInterface interface {
	Authenticate(cookie string) (userID uint, username string, err error)
	GetUserByID(userID uint) (*gdomain.User, error)
}

type Authenticator struct {
	redisClient *redis.Client
	Db          *gorm.DB
}

func NewAuthenticator(redisClient *redis.Client, Db *gorm.DB) *Authenticator {
	return &Authenticator{redisClient: redisClient, Db: Db}
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

	return session.UserID, session.Username, nil
}

func (a *Authenticator) GetUserByID(userID uint) (*gdomain.User, error) {
	var user gdomain.User
	err := a.Db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
