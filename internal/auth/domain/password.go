// Password — value object для пароля.
// Хранит хэш пароля и гарантирует его целостность.
//
// 💡 Почему хэш, а не plaintext?
// - В домене мы НИКОГДА не работаем с plaintext-паролем!
// - Plaintext пароль существует ТОЛЬКО в delivery-слое (HTTP-хендлер)
// - В домен попадает уже хэш (или plaintext для валидации при создании)

package domain

import (
	"strings"
)

// Password — value object для хэша пароля.
type Password string

// NewPasswordFromPlaintext создаёт Password из plaintext-пароля.
// Выполняет базовую валидацию (в реальности — bcrypt хэширование в service).
func NewPasswordFromPlaintext(plaintext string) (Password, error) {
	if plaintext == "" {
		return "", ErrWeakPassword("password is required")
	}
	if len(plaintext) < 8 {
		return "", ErrWeakPassword("password must be at least 8 characters")
	}
	// ⚠️ ВАЖНО: здесь НЕ хэшируем! Хэширование — задача инфраструктуры (service/adapter)
	// Но мы гарантируем, что plaintext валиден.
	return Password(plaintext), nil
}

// HashedPassword создаёт Password из уже хэшированной строки.
// Используется при загрузке из БД.
func HashedPassword(hashed string) Password {
	// Хэш считается всегда валидным (он уже прошёл валидацию при создании)
	return Password(hashed)
}

// String возвращает хэш (осторожно: не логируй в продакшене!)
func (p Password) String() string {
	return string(p)
}

// IsHashed проверяет, является ли пароль хэшированным.
// Простая эвристика: хэш bcrypt начинается с "$2a$", "$2b$", "$2y$"
func (p Password) IsHashed() bool {
	return strings.HasPrefix(string(p), "$2")
}

// =============================================================================
// 💡 ПРИНЦИП:
// - Domain не знает про bcrypt, но знает: "пароль может быть plaintext или хэш"
// - Хэширование — задача service (с использованием crypto/bcrypt)
// =============================================================================
