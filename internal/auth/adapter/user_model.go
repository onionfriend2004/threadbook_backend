// UserModel — это инфраструктурная модель для GORM.
// Это НЕ часть домена! Это "отражение" User в БД.

package adapter

import (
	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
)

// UserModel отображает User в таблицу БД.
type UserModel struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"` // храним как строку
	Password string // храним хэш как строку
}

// ToDomain конвертирует UserModel → domain.User.
func (m *UserModel) ToDomain() (*domain.User, error) {
	// Воссоздаём value objects
	email := domain.Email(m.Email)
	password := domain.HashedPassword(m.Password)

	user := &domain.User{
		ID:       m.ID,
		Email:    email,
		Password: password,
	}
	return user, nil
}

// FromDomain конвертирует domain.User → UserModel.
func FromDomain(u *domain.User) (*UserModel, error) {
	model := &UserModel{
		ID:       u.ID,
		Email:    u.Email.String(),
		Password: u.Password.String(),
	}
	return model, nil
}
