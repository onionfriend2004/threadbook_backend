package external

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type spoolRepo struct {
	db *gorm.DB
}

func NewSpoolRepo(db *gorm.DB) SpoolRepoInterface {
	return &spoolRepo{db: db}
}

// Проверка на нарушение уникальности (Postgres)
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Создание Spool + связь с владельцем
func (r *spoolRepo) CreateSpool(ctx context.Context, spool *gdomain.Spool, ownerID uint) (*gdomain.Spool, error) {
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

	userSpool := gdomain.UserSpool{
		UserID:  ownerID,
		SpoolID: spool.ID,
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

func (r *spoolRepo) GetSpoolByID(ctx context.Context, spoolID uint) (*gdomain.Spool, error) {
	var spool gdomain.Spool
	err := r.db.WithContext(ctx).First(&spool, spoolID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &spool, err
}

func (r *spoolRepo) UpdateSpool(ctx context.Context, spoolID uint, name, bannerLink string) (*gdomain.Spool, error) {
	var spool gdomain.Spool
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

func (r *spoolRepo) DeleteSpool(ctx context.Context, spoolID uint) error {
	return r.db.WithContext(ctx).Delete(&gdomain.Spool{}, spoolID).Error
}

func (r *spoolRepo) AddUserToSpoolByUsername(ctx context.Context, username string, spoolID uint) error {
	var user gdomain.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	userSpool := gdomain.UserSpool{
		UserID:  user.ID,
		SpoolID: spoolID,
	}

	if err := r.db.WithContext(ctx).Create(&userSpool).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyInSpool
		}
		return err
	}

	return nil
}

func (r *spoolRepo) RemoveUserFromSpool(ctx context.Context, userID, spoolID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND spool_id = ?", userID, spoolID).
		Delete(&gdomain.UserSpool{}).Error
}

func (r *spoolRepo) GetSpoolsByUser(ctx context.Context, userID uint) ([]gdomain.Spool, error) {
	var spools []gdomain.Spool
	err := r.db.WithContext(ctx).
		Joins("JOIN user_spools us ON us.spool_id = spools.id").
		Where("us.user_id = ?", userID).
		Find(&spools).Error
	return spools, err
}

func (r *spoolRepo) GetMembersBySpoolID(ctx context.Context, spoolID uint) ([]gdomain.User, error) {
	var users []gdomain.User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_spools us ON us.user_id = users.id").
		Where("us.spool_id = ?", spoolID).
		Find(&users).Error
	return users, err
}
