package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type UserRepoInterface interface {
	CreateUser(ctx context.Context, user gdomain.User) (*gdomain.User, error)

	GetUserByID(ctx context.Context, id uint) (*gdomain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*gdomain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*gdomain.User, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	VerifyUserEmail(ctx context.Context, userID uint) error
	// TODO: ExistsUsername Yes/No
}
