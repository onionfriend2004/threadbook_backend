package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type SpoolRepoInterface interface {
	CreateSpool(ctx context.Context, spool *gdomain.Spool, ownerID int) (*gdomain.Spool, error)
	GetSpoolByID(ctx context.Context, id uint) (*gdomain.Spool, error)
	UpdateSpool(ctx context.Context, spoolID int, name, bannerLink string) (*gdomain.Spool, error)
	DeleteSpool(ctx context.Context, id uint) error

	// join-таблицы spool <-> user
	AddUserToSpool(ctx context.Context, userID, spoolID int) error
	RemoveUserFromSpool(ctx context.Context, userID, spoolID int) error
	GetSpoolsByUser(ctx context.Context, userID int) ([]gdomain.Spool, error)
	GetMembersBySpoolID(ctx context.Context, spoolID int) ([]gdomain.User, error)
}
