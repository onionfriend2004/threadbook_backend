package external

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/domain"
)

type spoolRepo struct {
	db *gorm.DB
}

func NewSpoolRepo(db *gorm.DB) SpoolRepoInterface {
	return &spoolRepo{db: db}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *spoolRepo) CreateSpool(ctx context.Context, spool domain.Spool) (*domain.Spool, error) {
	if spool.Name == "" {
		return &domain.Spool{}, ErrInvalidSpool
	}
	err := r.db.WithContext(ctx).Create(&spool).Error
	if err != nil {
		if isUniqueViolation(err) {
			return &domain.Spool{}, ErrSpoolExists
		}
		return &domain.Spool{}, err
	}
	return &spool, nil
}

func (r *spoolRepo) GetSpoolByID(ctx context.Context, id uint) (*domain.Spool, error) {
	var spool domain.Spool
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&spool).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.Spool{}, ErrNotFound
	}
	return &spool, err
}

func (r *spoolRepo) UpdateSpool(ctx context.Context, spool domain.Spool) (*domain.Spool, error) {
	if spool.ID == 0 {
		return &domain.Spool{}, ErrInvalidSpool
	}
	err := r.db.WithContext(ctx).Save(&spool).Error
	return &spool, err
}

func (r *spoolRepo) DeleteSpool(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&domain.Spool{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *spoolRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Spool{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

var _ SpoolRepoInterface = (*spoolRepo)(nil)
