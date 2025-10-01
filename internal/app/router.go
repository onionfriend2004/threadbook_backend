package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/onionfriend2004/threadbook_backend/db"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/adapter"
	deliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/http"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run starts the HTTP server with graceful shutdown.
func Run(logger *zap.Logger) error {
	dbConn, err := db.GetPostgres()
	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		return err
	}

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID) // - RequestID: генерирует уникальный ID для каждого запроса (полезен для трассировки).
	r.Use(middleware.RealIP)    // - RealIP: извлекает реальный IP клиента из заголовков (X-Forwarded-For и др.).
	r.Use(middleware.Recoverer) // - Recoverer: перехватывает паники в обработчиках и предотвращает падение сервера.

	r.Mount("/api", apiRouter(dbConn, logger))

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Ожидаем отмены контекста (graceful shutdown будет в main, но можно и здесь)
	// В данном случае graceful shutdown лучше обрабатывать в main.go — так и оставим.
	// Эта функция просто запускает сервер и возвращает управление.

	return nil
}

func apiRouter(db *gorm.DB, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	// ===================== Auth =====================

	userRepo := adapter.NewGORMUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := deliveryHTTP.NewHandler(authService /*, logger.With(zap.String("component", "auth"))*/)

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	// ===================== Example =====================

	return r
}
