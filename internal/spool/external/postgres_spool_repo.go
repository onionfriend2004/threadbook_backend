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

func (r *spoolRepo) CreateSpool(ctx context.Context, spool *domain.Spool, ownerID int) (*domain.Spool, error) {
	if spool.Name == "" {
		return nil, ErrInvalidSpool
	}

	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(spool).Error; err != nil {
		tx.Rollback()
		if isUniqueViolation(err) {
			return nil, ErrSpoolExists
		}
		return nil, err
	}

	userSpool := domain.UserSpool{
		UserID:  ownerID,
		SpoolID: int(spool.ID),
	}
	if err := tx.Create(&userSpool).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return spool, nil
}

func (r *spoolRepo) GetSpoolByID(ctx context.Context, id uint) (*domain.Spool, error) {
	var spool domain.Spool
	err := r.db.WithContext(ctx).First(&spool, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &spool, err
}

func (r *spoolRepo) UpdateSpool(ctx context.Context, spoolID int, name, bannerLink string) (*domain.Spool, error) {
	var spool domain.Spool
	if err := r.db.WithContext(ctx).First(&spool, spoolID).Error; err != nil {
		return nil, err
	}

	if name != "" {
		spool.Name = name
	}
	if bannerLink != "" {
		spool.BannerLink = bannerLink
	}

	if err := r.db.WithContext(ctx).Save(&spool).Error; err != nil {
		return nil, err
	}
	return &spool, nil
}

func (r *spoolRepo) DeleteSpool(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Spool{}, id).Error
}

func (r *spoolRepo) AddUserToSpool(ctx context.Context, userID, spoolID int) error {
	us := domain.UserSpool{
		UserID:  userID,
		SpoolID: spoolID,
	}
	return r.db.WithContext(ctx).Create(&us).Error
}

func (r *spoolRepo) RemoveUserFromSpool(ctx context.Context, userID, spoolID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND spool_id = ?", userID, spoolID).
		Delete(&domain.UserSpool{}).Error
}

func (r *spoolRepo) GetSpoolsByUser(ctx context.Context, userID int) ([]domain.Spool, error) {
	var spools []domain.Spool
	err := r.db.WithContext(ctx).
		Joins("JOIN user_spools us ON us.spool_id = spools.id").
		Where("us.user_id = ?", userID).
		Find(&spools).Error
	return spools, err
}

func (r *spoolRepo) GetMembersBySpoolID(ctx context.Context, spoolID int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_spools us ON us.user_id = users.id").
		Where("us.spool_id = ?", spoolID).
		Find(&users).Error
	return users, err
}
