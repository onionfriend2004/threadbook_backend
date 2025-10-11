package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/onionfriend2004/threadbook_backend/config"
)

// MinioConnect создаёт подключение к MinIO и возвращает клиент.
// Оно не создаёт бакеты — только соединение.
func MinioConnect(cfg *config.Config) (*minio.Client, error) {
	endpoint := fmt.Sprintf("%s:%d", cfg.Minio.Host, cfg.Minio.Port)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	// Ping Pong
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets (check credentials or endpoint): %w", err)
	}

	return client, nil
}
