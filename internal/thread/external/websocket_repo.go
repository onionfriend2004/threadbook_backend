package external

import (
	"context"
	"time"
)

type WebsocketRepoInterface interface {
	PublishToUser(ctx context.Context, userID uint, data any) error
	PublishToThread(ctx context.Context, threadID uint, data any) error
	GenerateConnectToken(ctx context.Context, userID uint, ttl time.Duration) (string, error)
	GenerateSubscribeToken(ctx context.Context, userID uint, channel string, ttl time.Duration) (string, error)
}
