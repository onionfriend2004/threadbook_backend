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

func GetUserIDFromContext(ctx context.Context) (uint, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return 0, ErrNoUserIDInContext
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		if v < 0 {
			return 0, errors.New("negative user ID in context")
		}
		return uint(v), nil
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(id), nil
	case float64:
		if v < 0 {
			return 0, errors.New("negative user ID in context")
		}
		return uint(v), nil
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
