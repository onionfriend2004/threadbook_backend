package external

import (
	"context"
	"time"
)

type WebsocketRepoInterface interface {
	PublishToUser(ctx context.Context, userID uint, data any) error
	GenerateConnectToken(ctx context.Context, userID uint, ttl time.Duration) (string, error)
	GenerateSubscribeTokens(ctx context.Context, userID uint, threadIDs []uint, ttl time.Duration) (map[string]string, error)
}
