package external

import "context"

type SFUInterface interface {
	EnsureRoom(ctx context.Context, roomName string) error
}
