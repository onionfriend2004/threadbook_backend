package app

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/config"
	emailDeliveryNATS "github.com/onionfriend2004/threadbook_backend/internal/email/delivery/nats"
	emailExternal "github.com/onionfriend2004/threadbook_backend/internal/email/external"
	emailUsecase "github.com/onionfriend2004/threadbook_backend/internal/email/usecase"
	"go.uber.org/zap"
)

func initEmailConsumer(cfg *config.Config, nc *nats.Conn, logger *zap.Logger) emailDeliveryNATS.EmailConsumerInterface {
	emailRepo := emailExternal.NewMailRepository(
		cfg.Smtp.Server,
		cfg.Smtp.Port,
		cfg.Smtp.Username,
		cfg.Smtp.Password,
		cfg.Smtp.Sender,
	)
	emailUsecase := emailUsecase.NewEmailUsecase(emailRepo, logger.With(zap.String("service", "email")))
	return emailDeliveryNATS.NewEmailConsumer(
		nc,
		cfg.Nats.VerifyCodeSubject,
		emailUsecase,
		logger.With(zap.String("component", "email_consumer")),
	)
}

func startEmailConsumer(ctx context.Context, consumer emailDeliveryNATS.EmailConsumerInterface, logger *zap.Logger) {
	if err := consumer.Start(ctx); err != nil {
		logger.Error("email consumer failed", zap.Error(err))
	}
}
