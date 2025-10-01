package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/db"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/adapter"
	deliveryHTTP "github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/http"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/service"
	"gorm.io/gorm"
)

func Run() error {
	// Подключаемся к базе данных
	db, err := db.GetPostgres()
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	r.Mount("/api", apiRouter(db))

	httpServer := &http.Server{
		Addr:    ":8080", // добавил двоеточие
		Handler: r,
	}

	return httpServer.ListenAndServe()
}

func apiRouter(db *gorm.DB) chi.Router {
	r := chi.NewRouter()

	userRepo := adapter.NewGORMUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := deliveryHTTP.NewHandler(authService)

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	return r
}
