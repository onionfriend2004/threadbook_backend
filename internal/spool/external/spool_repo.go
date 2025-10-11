package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type SpoolRepoInterface interface {
	CreateSpool(ctx context.Context, spool *gdomain.Spool, ownerID uint) (*gdomain.Spool, error)
	GetSpoolByID(ctx context.Context, spoolID uint) (*gdomain.Spool, error)
	UpdateSpool(ctx context.Context, spoolID uint, name, bannerLink string) (*gdomain.Spool, error)
	DeleteSpool(ctx context.Context, spoolID uint) error

	AddUserToSpoolByUsername(ctx context.Context, username string, spoolID uint) error
	RemoveUserFromSpool(ctx context.Context, userID, spoolID uint) error
	GetSpoolsByUser(ctx context.Context, userID uint) ([]gdomain.Spool, error)
	GetMembersBySpoolID(ctx context.Context, spoolID uint) ([]gdomain.User, error)
}
