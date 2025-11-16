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

	// join-таблицы spool <-> user
	AddUserToSpoolByUsername(ctx context.Context, username string, spoolID uint) error
	RemoveUserFromSpool(ctx context.Context, userID, spoolID uint) error
	GetSpoolsByUser(ctx context.Context, userID uint) ([]gdomain.SpoolWithCreator, error)
	GetMembersBySpoolID(ctx context.Context, spoolID uint) ([]SpoolMember, error)

	IsUserInSpool(ctx context.Context, userID uint, spoolID uint) (bool, error)

	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error

	RemoveAllGuestsFromSpool(ctx context.Context, spoolID uint) error
	GetUserSpoolStatus(ctx context.Context, userID uint, spoolID uint) (*gdomain.UserSpool, error)
	GetUserSpoolStatusByUsername(ctx context.Context, username string, spoolID uint) (*gdomain.UserSpool, error)
	UpdateUserAccessLevel(ctx context.Context, userID uint, spoolID uint, accessLevel uint) error
}

type SpoolMember struct {
	gdomain.User
	AccessLevel uint `json:"access_level"`
}
