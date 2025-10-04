package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type SessionRepoInterface interface {
	GetSessionByID(ctx context.Context, session_id string) (*domain.Session, error)
	AddSessionForUser(ctx context.Context, user *gdomain.User) (*domain.Session, error)
	DelSessionByID(ctx context.Context, session_id string) error

	generateSessionID() (string, error)
}
