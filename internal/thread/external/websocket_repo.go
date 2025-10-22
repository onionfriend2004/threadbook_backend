package external

import (
	"context"
	"time"
)

type WebsocketRepoInterface interface {
	PublishToUser(ctx context.Context, userID uint, data any) error
	GenerateUserToken(ctx context.Context, userID uint, ttl time.Duration) (string, error)
}
