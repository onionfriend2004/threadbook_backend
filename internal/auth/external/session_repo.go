package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
)

type SessionRepoInterface interface {
	GetSessionByID(ctx context.Context, session_id string) (*domain.Session, error)
	AddSessionForUser(ctx context.Context, user *domain.User) (*domain.Session, error)
	DelSessionByID(ctx context.Context, session_id string) error

	generateSessionID() (string, error)
}
