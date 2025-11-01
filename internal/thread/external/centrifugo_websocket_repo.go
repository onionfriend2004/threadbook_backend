package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/centrifugal/gocent/v3"
	"github.com/golang-jwt/jwt/v5"
)

type websocketRepo struct {
	client      *gocent.Client
	secret      string // JWT secret
	tokenIssuer string
}

func NewWebsocketRepo(client *gocent.Client, secret, tokenIssuer string) WebsocketRepoInterface {
	return &websocketRepo{
		client:      client,
		secret:      secret,
		tokenIssuer: tokenIssuer,
	}
}

// user channel: "user#{id}"
func (r *websocketRepo) userChannel(userID uint) string {
	return fmt.Sprintf("user#%d", userID)
}

// thread channel: "thread#{id}"
func (r *websocketRepo) threadChannel(threadID uint) string {
	return fmt.Sprintf("thread#%d", threadID)
}

func (r *websocketRepo) PublishToUser(ctx context.Context, userID uint, data any) error {
	channel := r.userChannel(userID)

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal publish data: %w", err)
	}

	if _, err := r.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("centrifugo publish failed: %w", err)
	}

	return nil
}

func (r *websocketRepo) PublishToThread(ctx context.Context, threadID uint, data any) error {
	channel := r.threadChannel(threadID)

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal publish data: %w", err)
	}

	if _, err := r.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("centrifugo publish failed: %w", err)
	}

	return nil
}

// CONNECT JWT
func (r *websocketRepo) GenerateConnectToken(ctx context.Context, userID uint, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
	}
	if r.tokenIssuer != "" {
		claims["iss"] = r.tokenIssuer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(r.secret))
}

func (r *websocketRepo) GenerateSubscribeToken(ctx context.Context, userID uint, channel string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":     fmt.Sprintf("%d", userID),
		"channel": channel,
		"exp":     now.Add(ttl).Unix(),
		"iat":     now.Unix(),
	}

	if r.tokenIssuer != "" {
		claims["iss"] = r.tokenIssuer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(r.secret))
}
