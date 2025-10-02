package external

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
)

var (
	ErrInvalidUser = domain.ErrInvalidUser
	ErrUserExists  = errors.New("user already exists")
	ErrNotFound    = domain.ErrNotFound
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

func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.Email == "" || user.Username == "" || user.PasswordHash == "" {
		return domain.User{}, ErrInvalidUser
	}
	err := r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ErrUserExists
		}
		return domain.User{}, err
	}
	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uint) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	normalized := domain.NormalizeEmail(email)
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", normalized).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	normalized := domain.NormalizeUsername(username)
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", normalized).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}

func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	normalized := domain.NormalizeEmail(email)
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", normalized).Count(&count).Error
	return count > 0, err
}

func (r *userRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	normalized := domain.NormalizeUsername(username)
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("username = ?", normalized).Count(&count).Error
	return count > 0, err
}

var _ UserRepoInterface = (*userRepo)(nil)
