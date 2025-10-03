package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/redis/go-redis/v9"
)

// Подключение к Redis
func RedisConnect(cfg *config.Config) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// Ping Pong
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
