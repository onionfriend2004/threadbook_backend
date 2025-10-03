package app

import (
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"net/http"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/infra"
	deliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/http"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
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
	// ===================== OtherConn =====================

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID) // - RequestID: генерирует уникальный ID для каждого запроса (полезен для трассировки).
	r.Use(middleware.RealIP)    // - RealIP: извлекает реальный IP клиента из заголовков (X-Forwarded-For и др.).
	r.Use(middleware.Recoverer) // - Recoverer: перехватывает паники в обработчиках и предотвращает падение сервера.

	apiRouter, err := apiRouter(config, postgreConn, redisConn, logger)
	if err != nil {
		logger.Error("failed to create API router", zap.Error(err))
		return err
	}
	r.Mount("/api", apiRouter)

	httpServer := &http.Server{
		Addr:    config.App.Port,
		Handler: r,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	return nil
}

func apiRouter(cfg *config.Config, db *gorm.DB, redis *redis.Client, logger *zap.Logger) (chi.Router, error) {
	r := chi.NewRouter()

	// ===================== Auth =====================

	// external
	userRepo := external.NewUserRepo(db)
	sessionRepo := external.NewSessionRepo(redis, time.Duration(cfg.UserSession.TTL)*time.Second)

	// utils
	hasher, err := hasher.NewArgon2HasherFromConfig(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}
	cookieConfig := config.NewCookieConfig(cfg)

	// usecase
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, hasher, logger)

	// handler
	authHandler := deliveryHTTP.NewAuthHandler(authUsecase, logger.With(zap.String("component", "auth")), cookieConfig)
	authHandler.Routes(r)

	// ===================== Spool =====================

	// ===================== Thread =====================

	// ===================== Other =====================

	return r, nil
}
