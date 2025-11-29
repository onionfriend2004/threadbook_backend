package deliveryNATS

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		MaxMsgs:   -1,                  // Количество сообщений (ограничение)
		MaxAge:    7 * 24 * time.Hour,  // Храним сообщения неделю
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	sub, err := c.js.Subscribe(c.subject, c.handleMessage,
		nats.Durable(c.consumer),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.AckWait(60*time.Second),
		nats.MaxDeliver(3),
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
		msg.Term()
		return
	}

	var err error
	switch msg.Subject {
	case "user.registered":
		err = c.usecase.SendWelcomeEmail(&eventBody)
	case "user.code.resend":
		err = c.usecase.SendCodeResendEmail(&eventBody)
	default:
		c.logger.Warn("unknown subject, skipping", zap.String("subject", msg.Subject))
		msg.Term()
		return
	}

	if err != nil {
		if errors.Is(err, usecase.ErrPermanentEmailError) {
			c.logger.Warn("permanent email error, terminating message",
				zap.String("subject", msg.Subject),
				zap.String("email", eventBody.Email),
				zap.Error(err))
			msg.Term()
			return
		}

		c.logger.Error("temporary processing error, will retry",
			zap.String("subject", msg.Subject),
			zap.String("email", eventBody.Email),
			zap.Error(err))
		msg.NakWithDelay(10 * time.Second)
		return
	}

	if err := msg.Ack(); err != nil {
		c.logger.Warn("Failed to ack message — terminating to prevent redelivery",
			zap.String("subject", msg.Subject),
			zap.String("email", eventBody.Email),
			zap.Error(err))
		msg.Term()
	}
}
