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

// SUBSCRIBE JWT tokens for all user channels
func (r *websocketRepo) GenerateSubscribeTokens(ctx context.Context, userID uint, threadIDs []uint, ttl time.Duration) (map[string]string, error) {
	tokens := make(map[string]string)

	// user channel
	userCh := r.userChannel(userID)
	userToken, err := r.generateChannelToken(userID, userCh, ttl)
	if err != nil {
		return nil, err
	}
	tokens[userCh] = userToken

	// thread channels
	for _, threadID := range threadIDs {
		threadCh := r.threadChannel(threadID)
		token, err := r.generateChannelToken(userID, threadCh, ttl)
		if err != nil {
			return nil, err
		}
		tokens[threadCh] = token
	}

	return tokens, nil
}

// private helper to generate SUB JWT
func (r *websocketRepo) generateChannelToken(userID uint, channel string, ttl time.Duration) (string, error) {
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
