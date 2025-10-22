package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type MessageRepoInterface interface {
	Create(ctx context.Context, m *gdomain.Message) error
	CreateWithPayloads(ctx context.Context, m *gdomain.Message) error
	GetByThreadID(ctx context.Context, threadID uint, limit, offset int) ([]gdomain.Message, error)
	GetByID(ctx context.Context, id uint) (*gdomain.Message, error)
	DeleteByID(ctx context.Context, id uint) error
	CountByThreadID(ctx context.Context, threadID uint) (int64, error)
}
