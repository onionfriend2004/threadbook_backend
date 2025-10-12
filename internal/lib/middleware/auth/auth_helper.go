package auth

import (
	"context"
	"errors"
)

var (
	ErrNoUserIDInContext   = errors.New("no user ID in context")
	ErrNoUsernameInContext = errors.New("no username in context")
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

func GetUserIDFromContext(ctx context.Context) (uint, error) {
	if v, ok := ctx.Value(UserIDKey).(uint); ok {
		return v, nil
	}
	return 0, ErrNoUserIDInContext
}

func GetUsernameFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(UsernameKey)
	if value == nil {
		return "", ErrNoUsernameInContext
	}

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", errors.New("invalid username type in context")
	}
}
