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
	channelNS   string // например "user"
	secret      string // JWT secret
	tokenIssuer string // optional iss claim
}

func NewWebsocketRepo(client *gocent.Client, channelNS, secret, tokenIssuer string) WebsocketRepoInterface {
	return &websocketRepo{
		client:      client,
		channelNS:   channelNS,
		secret:      secret,
		tokenIssuer: tokenIssuer,
	}
}

// channelName возвращает имя пользовательского канала вида "user:{id}" или "namespace:user:{id}"
func (r *websocketRepo) channelName(userID uint) string {
	if r.channelNS == "" {
		return fmt.Sprintf("user:%d", userID)
	}
	return fmt.Sprintf("%s:user:%d", r.channelNS, userID)
}

// PublishToUser публикует сообщение конкретному пользователю
func (r *websocketRepo) PublishToUser(ctx context.Context, userID uint, data any) error {
	channel := r.channelName(userID)

	var payload []byte
	switch v := data.(type) {
	case []byte:
		payload = v
	case json.RawMessage:
		payload = []byte(v)
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal publish data: %w", err)
		}
		payload = b
	}

	_, err := r.client.Publish(ctx, channel, payload)
	if err != nil {
		return fmt.Errorf("centrifugo publish failed: %w", err)
	}
	return nil
}

// GenerateUserToken генерирует токен для подключения к пользовательскому каналу
func (r *websocketRepo) GenerateUserToken(ctx context.Context, userID uint, ttl time.Duration) (string, error) {
	now := time.Now()
	exp := now.Add(ttl).Unix()

	channel := r.channelName(userID)

	claims := jwt.MapClaims{
		"sub":     fmt.Sprintf("%d", userID),
		"channel": channel,
		"exp":     exp,
		"iat":     now.Unix(),
	}
	if r.tokenIssuer != "" {
		claims["iss"] = r.tokenIssuer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(r.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}
