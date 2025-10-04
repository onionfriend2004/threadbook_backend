package external

import (
	"context"
)

type VerifyCodeRepoInterface interface {
	SaveCode(ctx context.Context, userID uint, code int) error
	VerifyCode(ctx context.Context, userID uint, code int) (bool, error)

	GenerateCode() (int, error)
}
