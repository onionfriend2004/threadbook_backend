package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/domain"
)

type SpoolRepoInterface interface {
	CreateSpool(ctx context.Context, spool domain.Spool) (*domain.Spool, error)
	GetSpoolByID(ctx context.Context, id uint) (*domain.Spool, error)
	UpdateSpool(ctx context.Context, spool domain.Spool) (*domain.Spool, error)
	DeleteSpool(ctx context.Context, id uint) error
	ExistsByName(ctx context.Context, name string) (bool, error)

	// // нужно для юзкейсов
	// AddUserToSpool(ctx context.Context, userID, spoolID int) error
	// RemoveUserFromSpool(ctx context.Context, userID, spoolID int) error
	// GetUserSpools(ctx context.Context, userID int) ([]domain.Spool, error)
	// GetSpoolMembers(ctx context.Context, spoolID int) ([]domain.User, error)
}
