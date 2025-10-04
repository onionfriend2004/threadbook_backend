package infra

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/config"
)

func NatsConnect(cfg *config.Config) (*nats.Conn, error) {
	url := fmt.Sprintf("nats://%s:%d", cfg.Nats.Host, cfg.Nats.Port)

	nc, err := nats.Connect(url,
		nats.Name(cfg.Nats.Name),
		nats.Timeout(time.Duration(cfg.Nats.Timeout)*time.Second),
		nats.MaxReconnects(cfg.Nats.MaxReconnects),
		nats.ReconnectWait(time.Duration(cfg.Nats.ReconnectWait)*time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}
