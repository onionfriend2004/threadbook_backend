// helpers.go
package deliveryHTTP

import (
	"context"
	"errors"
	"strconv"
)

var (
	ErrNoUserIDInContext = errors.New("no user ID in context")
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
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
