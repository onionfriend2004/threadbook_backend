package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"net/http"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/infra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run starts the HTTP server with graceful shutdown.
func Run(config *config.Config, logger *zap.Logger) error {
	// ===================== PostgreConn =====================
	dbConn, err := infra.PostgresConnect(config)
	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		return err
	}
	// ===================== OtherConn =====================

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID) // - RequestID: генерирует уникальный ID для каждого запроса (полезен для трассировки).
	r.Use(middleware.RealIP)    // - RealIP: извлекает реальный IP клиента из заголовков (X-Forwarded-For и др.).
	r.Use(middleware.Recoverer) // - Recoverer: перехватывает паники в обработчиках и предотвращает падение сервера.

	r.Mount("/api", apiRouter(dbConn, logger))

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

func apiRouter(db *gorm.DB, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	// ===================== Auth =====================

	// init other...
	// authHandler := deliveryHTTP.NewHandler( /*, logger.With(zap.String("component", "auth"))*/ )

	// r.Post("/register", authHandler.Register)
	// r.Post("/login", authHandler.Login)

	// ===================== Spool =====================

	// ===================== Thread =====================

	// ===================== Other =====================

	return r
}
