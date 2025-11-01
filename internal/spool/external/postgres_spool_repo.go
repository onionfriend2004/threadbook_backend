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

func (r *spoolRepo) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx := r.db.Begin() // GORM
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, "tx", tx) // передаём tx через контекст

	err := fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Проверка существования пользователя по ID
func (r *spoolRepo) UserExistsByID(ctx context.Context, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&gdomain.User{}).
		Where("id = ?", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Создание Spool + связь с владельцем
func (r *spoolRepo) CreateSpool(ctx context.Context, spool *gdomain.Spool, ownerID uint) (*gdomain.Spool, error) {
	if spool.Name == "" {
		return nil, ErrInvalidSpool
	}

	// Проверяем, существует ли владелец
	exists, err := r.UserExistsByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
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
	// Проверяем, существует ли пользователь
	exists, err := r.UserExistsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	return r.db.WithContext(ctx).
		Where("user_id = ? AND spool_id = ?", userID, spoolID).
		Delete(&gdomain.UserSpool{}).Error
}

func (r *spoolRepo) GetSpoolsByUser(ctx context.Context, userID uint) ([]gdomain.SpoolWithCreator, error) {
	var result []gdomain.SpoolWithCreator

	err := r.db.WithContext(ctx).
		Table("spools").
		Select(`
			spools.id,
			spools.name,
			spools.banner_link,
			CASE WHEN spools.creator_id = ? THEN TRUE ELSE FALSE END AS is_creator
		`, userID).
		Joins("JOIN user_spools us ON us.spool_id = spools.id").
		Where("us.user_id = ?", userID).
		Scan(&result).Error

	return result, err
}

func (r *spoolRepo) GetMembersBySpoolID(ctx context.Context, spoolID uint) ([]gdomain.User, error) {
	var users []gdomain.User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_spools us ON us.user_id = users.id").
		Where("us.spool_id = ?", spoolID).
		Find(&users).Error
	return users, err
}

func (r *spoolRepo) IsUserInSpool(ctx context.Context, userID uint, spoolID uint) (bool, error) {
	// Проверяем, существует ли пользователь
	exists, err := r.UserExistsByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrUserNotFound
	}

	var count int64
	err = r.db.WithContext(ctx).
		Table("user_spools").
		Where("user_id = ? AND spool_id = ?", userID, spoolID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
