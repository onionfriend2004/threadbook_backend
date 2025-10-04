package external

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

type verifyCodeRepo struct {
	redisClient   *redis.Client
	verifyCodeTTL time.Duration
}

func NewVerifyCodeRepo(redisClient *redis.Client, verifyCodeTTL time.Duration) VerifyCodeRepoInterface {
	return &verifyCodeRepo{
		redisClient:   redisClient,
		verifyCodeTTL: verifyCodeTTL,
	}
}

func (r *verifyCodeRepo) SaveCode(ctx context.Context, userID uint, code int) error {
	key := fmt.Sprintf("verify_code:%d", userID)
	return r.redisClient.Set(ctx, key, code, time.Duration(r.verifyCodeTTL)*time.Second).Err()
}

func (r *verifyCodeRepo) VerifyCode(ctx context.Context, userID uint, code int) (bool, error) {
	key := fmt.Sprintf("verify_code:%d", userID)

	// Lua-скрипт для атомарной проверки + удаления
	script := redis.NewScript(`
		local stored = redis.call("GET", KEYS[1])
		if stored and tonumber(stored) == tonumber(ARGV[1]) then
			redis.call("DEL", KEYS[1])
			return 1
		end
		return 0
	`)

	result, err := script.Run(ctx, r.redisClient, []string{key}, code).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}

func (r *verifyCodeRepo) GenerateCode() (int, error) { // 6-значный
	max := big.NewInt(900000) // 999999 - 100000 + 1
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	code := int(n.Int64()) + 100000
	return code, nil
}

var _ VerifyCodeRepoInterface = (*verifyCodeRepo)(nil)
