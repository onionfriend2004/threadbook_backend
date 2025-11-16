package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type MessageRepoInterface interface {
	Create(ctx context.Context, m *gdomain.Message) error
	CreateWithPayloads(ctx context.Context, m *gdomain.Message) error
	GetByThreadCursor(ctx context.Context, threadID uint, cursorID uint, limit int, forward bool) ([]gdomain.Message, error)
	GetByID(ctx context.Context, id uint) (*gdomain.Message, error)
	Update(ctx context.Context, m *gdomain.Message) error
	DeleteByID(ctx context.Context, id uint) error
	CountByThreadID(ctx context.Context, threadID uint) (int64, error)
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}
