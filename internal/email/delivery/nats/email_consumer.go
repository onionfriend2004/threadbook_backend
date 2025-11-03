package deliveryNATS

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/internal/email/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"go.uber.org/zap"
)

type EmailConsumerInterface interface {
	Start(ctx context.Context) error
	Stop()
	handleMessage(msg *nats.Msg)
}

type emailConsumer struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	subject  string
	stream   string
	consumer string
	usecase  usecase.EmailUsecaseInterface
	logger   *zap.Logger
	quitChan chan struct{}
	sub      *nats.Subscription
}

func NewEmailConsumer(
	nc *nats.Conn,
	subject string,
	stream string,
	consumerName string,
	usecase usecase.EmailUsecaseInterface,
	logger *zap.Logger,
) (EmailConsumerInterface, error) {
	// Получаем JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	return &emailConsumer{
		nc:       nc,
		js:       js,
		subject:  subject,
		stream:   stream,
		consumer: consumerName,
		usecase:  usecase,
		logger:   logger,
		quitChan: make(chan struct{}),
	}, nil
}

func (c *emailConsumer) Start(ctx context.Context) error {
	// Создаем stream если не существует с WILDCARD subject
	_, err := c.js.AddStream(&nats.StreamConfig{
		Name:      c.stream,
		Subjects:  []string{c.subject}, // "user.*" - wildcard для всех user событий
		Retention: nats.LimitsPolicy,   // Для broadcast
		MaxMsgs:   10000,               // Ограничиваем количество сообщений
		MaxAge:    24 * time.Hour,      // Храним сообщения 24 часа
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Подписываемся на WILDCARD subject
	sub, err := c.js.Subscribe(c.subject, c.handleMessage,
		nats.Durable(c.consumer),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.AckWait(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to JetStream: %w", err)
	}

	c.sub = sub
	c.logger.Info("email JetStream consumer started",
		zap.String("subject", c.subject),
		zap.String("stream", c.stream),
		zap.String("consumer", c.consumer))

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
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	close(c.quitChan)
}

func (c *emailConsumer) handleMessage(msg *nats.Msg) {
	var eventBody gdomain.UserRegisteredEvent
	if err := json.Unmarshal(msg.Data, &eventBody); err != nil {
		c.logger.Error("failed to unmarshal email event", zap.Error(err))
		msg.Nak()
		return
	}

	// Определяем тип события по subject
	switch msg.Subject {
	case "user.registered":
		if eventBody.Type != event.UserRegistered {
			c.logger.Warn("event type mismatch for subject",
				zap.String("subject", msg.Subject),
				zap.Int("event_type", eventBody.Type))
		}
		if err := c.usecase.SendWelcomeEmail(&eventBody); err != nil {
			c.logger.Error("failed to process welcome email event",
				zap.String("subject", msg.Subject),
				zap.Int("type", eventBody.Type),
				zap.String("email", eventBody.Email),
				zap.Error(err))
			msg.Nak()
			return
		}

	case "user.code.resend":
		if eventBody.Type != event.UserRequestResendVerifyCode {
			c.logger.Warn("event type mismatch for subject",
				zap.String("subject", msg.Subject),
				zap.Int("event_type", eventBody.Type))
		}
		if err := c.usecase.SendCodeResendEmail(&eventBody); err != nil {
			c.logger.Error("failed to process code resend email event",
				zap.String("subject", msg.Subject),
				zap.Int("type", eventBody.Type),
				zap.String("email", eventBody.Email),
				zap.Error(err))
			msg.Nak()
			return
		}

	default:
		c.logger.Warn("unknown subject, skipping", zap.String("subject", msg.Subject))
		msg.Nak()
		return
	}

	// Подтверждаем обработку
	if err := msg.Ack(); err != nil {
		c.logger.Error("failed to ack message", zap.Error(err))
	}
}
