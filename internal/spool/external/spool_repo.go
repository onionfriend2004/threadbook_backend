package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/domain"
)

type SpoolRepoInterface interface {
	CreateSpool(ctx context.Context, spool *domain.Spool, ownerID int) (*domain.Spool, error)
	GetSpoolByID(ctx context.Context, id uint) (*domain.Spool, error)
	UpdateSpool(ctx context.Context, spoolID int, name, bannerLink string) (*domain.Spool, error)
	DeleteSpool(ctx context.Context, id uint) error

	// join-таблицы spool <-> user
	AddUserToSpool(ctx context.Context, userID, spoolID int) error
	RemoveUserFromSpool(ctx context.Context, userID, spoolID int) error
	GetSpoolsByUser(ctx context.Context, userID int) ([]domain.Spool, error)
	GetMembersBySpoolID(ctx context.Context, spoolID int) ([]domain.User, error)
}
