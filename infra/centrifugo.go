package infra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/centrifugal/gocent/v3"
	"github.com/onionfriend2004/threadbook_backend/config"
)

// CentrifugoConnect подключается к Centrifugo API.
func CentrifugoConnect(cfg *config.Config) (*gocent.Client, error) {
	scheme := "http"
	if cfg.Centrifugo.UseSSL {
		scheme = "https"
	}

	apiURL := fmt.Sprintf("%s://%s:%d/api", scheme, cfg.Centrifugo.Host, cfg.Centrifugo.Port)
	client := gocent.New(gocent.Config{
		Addr: apiURL,
		Key:  cfg.Centrifugo.APIKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second, // Устанавливаем таймаут
		},
	})

	_, err := client.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("centrifugo connection check failed: %w", err)
	}

	return client, nil
}
