// User — это AGGREGATE ROOT (корень агрегата) в bounded context "Auth".
//
// 🔑 Свойства Aggregate Root:
// - Является точкой входа для всех операций с агрегатом
// - Гарантирует целостность всего агрегата
// - Внешний мир взаимодействует ТОЛЬКО с корнем (не с Email/Password напрямую)
// - Содержит бизнес-логику как методы (богатая модель!)

package domain

import "errors"

// User представляет пользователя в системе.
// Это агрегат, состоящий из:
// - ID (идентичность агрегата)
// - Email (value object)
// - Password (value object)
type User struct {
	ID       string
	Email    Email
	Password Password
}

// NewUser создаёт нового пользователя.
// Принимает уже валидные Email и Password (plaintext).
//
// 💡 Почему фабрика в домене?
// - Гарантирует, что User создаётся в валидном состоянии
// - Инкапсулирует правила создания
func NewUser(email Email, password Password) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	// Генерация ID — задача инфраструктуры! Здесь ID остаётся пустым.
	// Агрегат будет сохранён через репозиторий, который присвоит ID.
	return &User{
		Email:    email,
		Password: password,
	}, nil
}

// ChangeEmail позволяет изменить email с валидацией.
// Это бизнес-операция, поэтому она — метод агрегата.
func (u *User) ChangeEmail(newEmail Email) error {
	if newEmail == "" {
		return errors.New("new email is required")
	}
	u.Email = newEmail
	return nil
}

// VerifyPassword проверяет, соответствует ли plaintext-пароль хэшу.
// ⚠️ ВАЖНО: эта логика — в домене, потому что это бизнес-правило!
// Но реализация сравнения — в инфраструктуре (bcrypt).
// Поэтому мы передаём функцию сравнения как зависимость.
//
// Однако для простоты в этом примере предположим, что Password уже знает,
// является ли он хэшем, и сравнение делается вне домена (в service).
// Это компромисс ради читаемости.
//
// В идеале: домен определяет интерфейс PasswordHasher, а service его реализует.
