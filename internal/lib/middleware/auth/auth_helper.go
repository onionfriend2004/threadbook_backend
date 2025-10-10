package auth

import (
	"context"
	"errors"
	"strconv"
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

func GetUserIDFromContext(ctx context.Context) (int, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return 0, ErrNoUserIDInContext
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	case float64:
		return int(v), nil
	default:
		return 0, errors.New("invalid user ID type in context")
	}
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
