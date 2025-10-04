package deliveryNATS

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/internal/email/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"go.uber.org/zap"
)

type EmailConsumerInterface interface {
	Start(ctx context.Context) error
	Stop()
	handleMessage(msg *nats.Msg)
}

type emailConsumer struct {
	nc       *nats.Conn
	subject  string
	usecase  usecase.EmailUsecaseInterface
	logger   *zap.Logger
	quitChan chan struct{}
}

func NewEmailConsumer(
	nc *nats.Conn,
	subject string,
	usecase usecase.EmailUsecaseInterface,
	logger *zap.Logger,
) EmailConsumerInterface {
	return &emailConsumer{
		nc:       nc,
		subject:  subject,
		usecase:  usecase,
		logger:   logger,
		quitChan: make(chan struct{}),
	}
}

// Блокирующий метод — запускать в отдельной горутине.
func (c *emailConsumer) Start(ctx context.Context) error {
	sub, err := c.nc.Subscribe(c.subject, c.handleMessage)
	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS subject %s: %w", c.subject, err)
	}
	defer sub.Unsubscribe()

	c.logger.Info("email consumer started", zap.String("subject", c.subject))

	select {
	case <-ctx.Done():
		c.logger.Info("email consumer stopped by context")
		return nil
	case <-c.quitChan:
		c.logger.Info("email consumer stopped manually")
		return nil
	}
}

func (c *emailConsumer) Stop() {
	close(c.quitChan)
}

func (c *emailConsumer) handleMessage(msg *nats.Msg) {
	var event gdomain.EmailEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		c.logger.Error("failed to unmarshal email event", zap.Error(err))
		return
	}

	if err := c.usecase.SendMessageOnEmail(&event); err != nil {
		c.logger.Error("failed to process email event",
			zap.Int("type", event.Type),
			zap.String("email", event.Email),
			zap.Error(err))
	}
}

var _ EmailConsumerInterface = (*emailConsumer)(nil)
