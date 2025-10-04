package external

import (
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

// USER_REPO ERRORS
var (
	ErrInvalidUser  = gdomain.ErrInvalidUser
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = gdomain.ErrNotFound
	ErrNotFound     = gdomain.ErrNotFound
)

// SESSION_REPO ERRORS
var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidSessionData = errors.New("invalid session data")
)

// SEND_CODE_REPO ERRORS
var (
	ErrFailedToSendCode = errors.New("failed to send verification code")
)
