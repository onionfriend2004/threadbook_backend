// Package domain определяет бизнес-ошибки как часть языка предметной области.
//
// 🔑 Принципы:
// - Ошибки — часть домена, а не инфраструктуры
// - Каждая ошибка имеет КОД (для логики) и СООБЩЕНИЕ (для пользователя)
// - В delivery-слое ошибки маппятся на HTTP-статусы

package domain

import (
	"errors"
	"fmt"
)

// AuthErrorCode — коды ошибок авторизации.
type AuthErrorCode string

const (
	ErrCodeUserNotFound       AuthErrorCode = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists  AuthErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeInvalidCredentials AuthErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInvalidEmail       AuthErrorCode = "INVALID_EMAIL"
	ErrCodeWeakPassword       AuthErrorCode = "WEAK_PASSWORD"
)

// AuthError — структурированная бизнес-ошибка.
type AuthError struct {
	Code    AuthErrorCode
	Message string
}

// Error реализует error interface.
func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Конструкторы ошибок
var (
	ErrUserNotFound = &AuthError{
		Code:    ErrCodeUserNotFound,
		Message: "user not found",
	}
	ErrUserAlreadyExists = &AuthError{
		Code:    ErrCodeUserAlreadyExists,
		Message: "user already exists",
	}
	ErrInvalidCredentials = &AuthError{
		Code:    ErrCodeInvalidCredentials,
		Message: "invalid credentials",
	}
	ErrInvalidEmail = func(msg string) *AuthError {
		return &AuthError{Code: ErrCodeInvalidEmail, Message: msg}
	}
	ErrWeakPassword = func(msg string) *AuthError {
		return &AuthError{Code: ErrCodeWeakPassword, Message: msg}
	}
)

// IsAuthError проверяет, является ли ошибка AuthError.
func IsAuthError(err error) bool {
	var e *AuthError
	return errors.As(err, &e)
}
