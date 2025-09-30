package adapter

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"

	"gorm.io/gorm"
)

// GORMUserRepository реализует порт domain.UserRepository.
type GORMUserRepository struct {
	db *gorm.DB
}

func NewGORMUserRepository(db *gorm.DB) *GORMUserRepository {
	return &GORMUserRepository{db: db}
}

func (r *GORMUserRepository) Save(ctx context.Context, user *domain.User) error {
	model, err := FromDomain(user)
	if err != nil {
		return err
	}
	if model.ID == "" {
		model.ID = generateID()
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GORMUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email.String()).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func generateID() string {
	return "temp-id"
}
