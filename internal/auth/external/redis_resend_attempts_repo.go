package external

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	attemptKeyPrefix = "resend_attempts:user:"
	defaultTTL       = 1 * time.Hour // дефолт - 1 час
)

type AttemptSendCodeRedisRepo struct {
	client redis.UniversalClient
	ttl    time.Duration
}

func NewAttemptSendCodeRedisRepo(client redis.UniversalClient, ttl time.Duration) AttemptSendCodeRepoInterface {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &AttemptSendCodeRedisRepo{
		client: client,
		ttl:    ttl,
	}
}

func (r *AttemptSendCodeRedisRepo) key(userID uint) string {
	return attemptKeyPrefix + strconv.FormatUint(uint64(userID), 10)
}

func (r *AttemptSendCodeRedisRepo) GetSendAttempts(ctx context.Context, userID uint) (int, error) {
	key := r.key(userID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Ключ не существует — попыток ещё не было
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get attempts from redis: %w", err)
	}

	attempts, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid attempts value in redis: %w", err)
	}

	return attempts, nil
}

func (r *AttemptSendCodeRedisRepo) IncrementSendAttempts(ctx context.Context, userID uint) error {
	key := r.key(userID)

	// INCR создаёт ключ, если его нет, и увеличивает значение на 1
	err := r.client.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to increment attempts in redis: %w", err)
	}

	// ВСЕГДА обновляем TTL — счётчик живёт r.ttl после КАЖДОЙ попытки
	err = r.client.Expire(ctx, key, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for attempts key: %w", err)
	}

	return nil
}

func (r *AttemptSendCodeRedisRepo) ResetSendAttempts(ctx context.Context, userID uint) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

var _ AttemptSendCodeRepoInterface = (*AttemptSendCodeRedisRepo)(nil)
