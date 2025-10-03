package external

import (
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
)

// USER_REPO ERRORS
var (
	ErrInvalidUser  = domain.ErrInvalidUser
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = domain.ErrNotFound
	ErrNotFound     = domain.ErrNotFound
)

// SESSION_REPO ERRORS
var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidSessionData = errors.New("invalid session data")
)
