package usecase

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCodeIncorrect      = errors.New("invalid verify code")
	ErrTooManyAttempts    = errors.New("too many attempts to send")
	ErrAlreadyConfirmed   = errors.New("user email already verified")
)
