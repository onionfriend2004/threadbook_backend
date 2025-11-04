package external

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	gen "github.com/onionfriend2004/threadbook_backend/internal/lib/username_generator"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepoInterface {
	return &userRepo{db: db}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *userRepo) CreateUser(ctx context.Context, user gdomain.User) (*gdomain.User, error) {
	if user.Email == "" || user.Username == "" || user.PasswordHash == "" {
		return &gdomain.User{}, ErrInvalidUser
	}
	user.EmailVerify = false
	err := r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		if isUniqueViolation(err) {
			return &gdomain.User{}, ErrUserExists
		}
		return &gdomain.User{}, err
	}
	return &user, nil
}

// ============= NoAuth ===============

func (r *userRepo) generateUniqueCredentials(ctx context.Context) (string, string, error) {
	for i := 0; i < 5; i++ {
		username := gen.GenerateRandomUsername()
		email := fmt.Sprintf("guest.%s@noauth.temp", uuid.New().String())

		var count int64
		err := r.db.WithContext(ctx).Model(&gdomain.User{}).
			Where("username = ? OR email = ?", username, email).
			Count(&count).Error

		if err != nil {
			return "", "", err
		}

		if count == 0 {
			return username, email, nil
		}
	}
	return "", "", fmt.Errorf("failed to generate unique credentials")
}

func (r *userRepo) CreateNoAuthUser(ctx context.Context) (*gdomain.User, error) {
	username, email, err := r.generateUniqueCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate username: %w", err)
	}
	user := gdomain.User{
		Username:     username,
		Email:        email,
		PasswordHash: "",
		IsGuest:      true,
	}

	err = r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) UpgradeGuestToUser(ctx context.Context, user gdomain.User) (*gdomain.User, error) {
	if user.Email == "" || user.Username == "" || user.PasswordHash == "" {
		return nil, ErrInvalidUser
	}

	// Обновляем только нужные поля
	updateData := map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"email_verify":  false,
		"is_guest":      false, // меняем на полноценного пользователя
		"updated_at":    time.Now(),
	}

	// Выполняем UPDATE только для указанных полей
	result := r.db.WithContext(ctx).
		Model(&gdomain.User{}).
		Where("id = ?", user.ID).
		Updates(updateData)

	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, ErrUserExists
		}
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// Возвращаем обновленного пользователя
	var updatedUser gdomain.User
	if err := r.db.WithContext(ctx).First(&updatedUser, user.ID).Error; err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

// ========== НЕ NoAuth ===========

func (r *userRepo) GetUserByID(ctx context.Context, id uint) (*gdomain.User, error) {
	var user gdomain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &gdomain.User{}, ErrNotFound
	}
	return &user, err
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*gdomain.User, error) {
	normalized := gdomain.NormalizeEmail(email)
	var user gdomain.User
	err := r.db.WithContext(ctx).Where("email = ?", normalized).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &gdomain.User{}, ErrNotFound
	}
	return &user, err
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*gdomain.User, error) {
	normalized := gdomain.NormalizeUsername(username)
	var user gdomain.User
	err := r.db.WithContext(ctx).Where("username = ?", normalized).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &gdomain.User{}, ErrNotFound
	}
	return &user, err
}

func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	normalized := gdomain.NormalizeEmail(email)
	var count int64
	err := r.db.WithContext(ctx).Model(&gdomain.User{}).Where("email = ?", normalized).Count(&count).Error
	return count > 0, err
}

func (r *userRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	normalized := gdomain.NormalizeUsername(username)
	var count int64
	err := r.db.WithContext(ctx).Model(&gdomain.User{}).Where("username = ?", normalized).Count(&count).Error
	return count > 0, err
}

func (r *userRepo) VerifyUserEmail(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).
		Model(&gdomain.User{}).
		Where("id = ?", userID).
		Update("email_verify", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepo) GetUserProfileByUserID(ctx context.Context, userID uint) (*gdomain.Profile, error) {
	var profile gdomain.Profile

	err := r.db.WithContext(ctx).
		Model(&gdomain.Profile{}).
		Where("user_id = ?", userID).
		First(&profile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Возвращаем пустой профиль вместо ошибки
			return &gdomain.Profile{
				UserID:     int(userID),
				Nickname:   "",
				AvatarLink: "",
			}, nil
		}
		return nil, err
	}

	return &profile, nil
}

var _ UserRepoInterface = (*userRepo)(nil)
