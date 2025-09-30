// UserRepository — это PORT (порт) из Hexagonal Architecture.
// Объявлен в домене, реализован в инфраструктуре (adapter/).
//
// 🔑 Принципы:
// - Интерфейс зависит ТОЛЬКО от типов домена (User, Email)
// - Не знает про БД, GORM, SQL
// - Методы выражены на языке бизнеса

package domain

import "context"

// UserRepository определяет контракт для работы с пользователями.
type UserRepository interface {
	// FindByEmail ищет пользователя по email.
	// Возвращает ошибку, если не найден.
	FindByEmail(ctx context.Context, email Email) (*User, error)

	// Save сохраняет пользователя.
	// Присваивает ID, если он пустой (реализация в адаптере).
	Save(ctx context.Context, user *User) error
}
