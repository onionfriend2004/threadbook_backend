package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/centrifugal/gocent/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	livekit "github.com/livekit/server-sdk-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"net/http"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/infra"
	authDeliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/http"
	authExternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
	authUsecase "github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	fileDeliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/file/delivery/http"
	fileExternal "github.com/onionfriend2004/threadbook_backend/internal/file/external"
	fileUsecase "github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	spoolDeliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/http"
	spoolExternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	spoolUsecase "github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	threadDeliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/http"
	threadExternal "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	threadUsecase "github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run starts the HTTP server with graceful shutdown.
func Run(config *config.Config, logger *zap.Logger) error {
	// ===================== PostgreConn =====================
	postgreConn, err := infra.PostgresConnect(config)
	if err != nil {
		logger.Error("failed to connect to postgres", zap.Error(err))
		fmt.Print(config.Postgres)
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

	// ===================== CentrifugoConn =====================
	centrifugoClient, err := infra.CentrifugoConnect(config)
	if err != nil {
		logger.Error("failed to connect to Centrifugo", zap.Error(err))
		return err
	}

	// ===================== LiveKitConn =====================
	liveKitConn := infra.LiveKitConnect(config)

	// ===================== MinioConn =====================
	minioConn, err := infra.MinioConnect(config)
	if err != nil {
		logger.Error("failed to connect to MinIO", zap.Error(err))
		return err
	}

	// ===================== Email Consumer =====================
	emailConsumer := initEmailConsumer(config, natsConn, logger)

	// Создаём общий контекст для всего приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startEmailConsumer(ctx, emailConsumer, logger)

	// ===================== HTTP Server =====================
	r := chi.NewRouter()

	// Middlewares
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   config.CORS.AllowedOrigins,
		AllowedMethods:   config.CORS.AllowedMethods,
		AllowedHeaders:   config.CORS.AllowedHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
		MaxAge:           config.CORS.MaxAge,
	})

	r.Use(corsMiddleware.Handler) // CORS
	r.Use(middleware.RequestID)   // - RequestID: генерирует уникальный ID для каждого запроса (полезен для трассировки).
	r.Use(middleware.RealIP)      // - RealIP: извлекает реальный IP клиента из заголовков (X-Forwarded-For и др.).
	r.Use(middleware.Recoverer)   // - Recoverer: перехватывает паники в обработчиках и предотвращает падение сервера.

	apiRouter, err := apiRouter(config, postgreConn, redisConn, natsConn, liveKitConn, minioConn, centrifugoClient, logger)
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

func apiRouter(cfg *config.Config, db *gorm.DB, redis *redis.Client, nts *nats.Conn, livekit *livekit.RoomServiceClient, minio *minio.Client, centrifugo *gocent.Client, logger *zap.Logger) (chi.Router, error) {
	r := chi.NewRouter()
	// ===================== Auth =====================

	authenticator := auth.NewAuthenticator(redis)

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
	fileConfig := config.NewFileConfig(cfg)
	// usecase
	authauthUsecase := authUsecase.NewAuthUsecase(userRepo, sessionRepo, sendCodeRepo, verifyCodeRepo, hasher, logger)

	// handler
	authHandler := authDeliveryHTTP.NewAuthHandler(authauthUsecase, logger.With(zap.String("component", "auth")), cookieConfig)
	authHandler.Routes(r)

	// ===================== File =====================
	fileRepo := fileExternal.NewFileRepo(minio, cfg.Minio.Bucket)
	fileUC := fileUsecase.NewFileUsecase(fileRepo, logger)
	fileHandler := fileDeliveryHTTP.NewFileHandler(fileUC, logger)
	fileHandler.Routes(r)

	// ===================== Spool =====================
	spoolRepo := spoolExternal.NewSpoolRepo(db)
	spoolUC := spoolUsecase.NewSpoolUsecase(spoolRepo, fileUC, logger)
	spoolHandler := spoolDeliveryHTTP.NewSpoolHandler(spoolUC, logger, fileConfig)
	spoolHandler.Routes(r, authenticator)

	// ===================== Thread =====================
	// external repos
	threadRepo := threadExternal.NewThreadRepo(db, logger)
	liveKitRepo := threadExternal.NewLiveKitRepo(livekit, cfg.Room.EmptyTTL, cfg.Room.MaxParticipants)
	websocketRepo := threadExternal.NewWebsocketRepo(
		centrifugo,               // *gocent.Client
		cfg.Centrifugo.TokenHMAC, // JWT secret
		"threadbook",             // token issuer
	)
	// messages repo
	messageRepo := threadExternal.NewMessageRepo(db)

	// usecases
	threadUC := threadUsecase.NewThreadUsecase(threadRepo, logger)
	messageUC := threadUsecase.NewMessageUsecase(messageRepo, websocketRepo, threadRepo, time.Duration(cfg.Centrifugo.TTL)*time.Second, logger)
	roomUC := threadUsecase.NewRoomUsecase(threadRepo, liveKitRepo, cfg.LiveKit.URL, cfg.LiveKit.APIKey, cfg.LiveKit.APISecret, logger)

	// handler
	threadHandler := threadDeliveryHTTP.NewThreadHandler(threadUC, messageUC, roomUC, logger)
	threadHandler.Routes(r, authenticator)
	// ===================== Other =====================

	return r, nil
}
