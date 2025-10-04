package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"net/http"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/infra"
	authDeliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/http"
	authExternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
	authUsecase "github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run starts the HTTP server with graceful shutdown.
func Run(config *config.Config, logger *zap.Logger) error {
	// ===================== PostgreConn =====================
	postgreConn, err := infra.PostgresConnect(config)
	if err != nil {
		logger.Error("failed to connect to postgres", zap.Error(err))
		return err
	}
	// ===================== RedisConn =====================
	redisConn, err := infra.RedisConnect(config)
	if err != nil {
		logger.Error("failed to connect to redis", zap.Error(err))
		return err
	}
	// ===================== NatsConn =====================
	natsConn, err := infra.NatsConnect(config)
	if err != nil {
		logger.Error("failed to connect to NATS", zap.Error(err))
		return err
	}

	// ===================== Email Consumer =====================
	// emailConsumer := initEmailConsumer(config, natsConn, logger)

	// Создаём общий контекст для всего приложения
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// go startEmailConsumer(ctx, emailConsumer, logger)

	// ===================== HTTP Server =====================
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID) // - RequestID: генерирует уникальный ID для каждого запроса (полезен для трассировки).
	r.Use(middleware.RealIP)    // - RealIP: извлекает реальный IP клиента из заголовков (X-Forwarded-For и др.).
	r.Use(middleware.Recoverer) // - Recoverer: перехватывает паники в обработчиках и предотвращает падение сервера.

	apiRouter, err := apiRouter(config, postgreConn, redisConn, natsConn, logger)
	if err != nil {
		return err
	}
	r.Mount("/api", apiRouter)

	httpServer := &http.Server{
		Addr:    config.App.Port,
		Handler: r,
	}

	// Запускаем HTTP сервер
	go func() {
		logger.Info("starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down gracefully...")

	cancel()

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	logger.Info("server exited")
	return nil
}

func apiRouter(cfg *config.Config, db *gorm.DB, redis *redis.Client, nts *nats.Conn, logger *zap.Logger) (chi.Router, error) {
	r := chi.NewRouter()

	// ===================== Auth =====================

	// external
	userRepo := authExternal.NewUserRepo(db)
	sessionRepo := authExternal.NewSessionRepo(redis, time.Duration(cfg.UserSession.TTL)*time.Minute)
	sendCodeRepo := authExternal.NewSendCodeRepo(nts, cfg.Nats.VerifyCodeSubject)
	verifyCodeRepo := authExternal.NewVerifyCodeRepo(redis, time.Duration(cfg.VerifyCode.TTL)*time.Minute)

	// utils
	hasher, err := hasher.NewArgon2HasherFromConfig(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}
	cookieConfig := config.NewCookieConfig(cfg)

	// usecase
	authauthUsecase := authUsecase.NewAuthUsecase(userRepo, sessionRepo, sendCodeRepo, verifyCodeRepo, hasher, logger)

	// handler
	authHandler := authDeliveryHTTP.NewAuthHandler(authauthUsecase, logger.With(zap.String("component", "auth")), cookieConfig)
	authHandler.Routes(r)

	// ===================== Spool =====================

	// ===================== Thread =====================

	// ===================== Other =====================

	return r, nil
}
