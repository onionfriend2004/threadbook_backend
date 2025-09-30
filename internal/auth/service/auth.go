// Package service содержит APPLICATION SERVICES — сценарии использования.
//
// 🔑 Роль Application Service:
// - Оркестрирует доменные объекты для выполнения use case
// - Координирует работу с репозиториями, внешними сервисами
// - Не содержит бизнес-логики! (она в агрегатах)
// - Зависит от домена, но не от инфраструктуры (кроме интерфейсов)

package service

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"

	"golang.org/x/crypto/bcrypt"
)

// AuthService реализует сценарии авторизации.
type AuthService struct {
	userRepo domain.UserRepository
}

// NewAuthService создаёт новый сервис.
func NewAuthService(userRepo domain.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register регистрирует нового пользователя.
// Это use case: "зарегистрировать пользователя".
func (s *AuthService) Register(ctx context.Context, emailStr, passwordStr string) (string, error) {
	// 1. Создаём value objects из входных данных
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return "", err
	}

	plaintextPassword, err := domain.NewPasswordFromPlaintext(passwordStr)
	if err != nil {
		return "", err
	}

	// 2. Проверяем, не существует ли пользователь
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return "", domain.ErrUserAlreadyExists
	}

	// 3. Хэшируем пароль (инфраструктурная логика!)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword.String()), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hashedPassword := domain.HashedPassword(string(hashedBytes))

	// 4. Создаём агрегат
	user, err := domain.NewUser(email, hashedPassword)
	if err != nil {
		return "", err
	}

	// 5. Сохраняем через репозиторий
	if err := s.userRepo.Save(ctx, user); err != nil {
		return "", err
	}

	return user.ID, nil
}

// Login выполняет вход пользователя.
func (s *AuthService) Login(ctx context.Context, emailStr, passwordStr string) error {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	// Сравниваем plaintext-пароль с хэшем
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String()), []byte(passwordStr)); err != nil {
		return domain.ErrInvalidCredentials
	}

	return nil
}
