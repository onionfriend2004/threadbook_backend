package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
)

type UserRepoInterface interface {
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)

	GetUserByID(ctx context.Context, id uint) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// TODO: ExistsUsername Yes/No
}
