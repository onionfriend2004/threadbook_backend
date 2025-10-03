package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"go.uber.org/zap"
)

type AuthHandler struct {
	usecase      usecase.AuthUsecaseInterface
	logger       *zap.Logger
	cookieConfig *config.CookieConfig
}

func NewAuthHandler(
	usecase usecase.AuthUsecaseInterface,
	logger *zap.Logger,
	cookieConfig *config.CookieConfig,
) *AuthHandler {
	return &AuthHandler{
		usecase:      usecase,
		logger:       logger,
		cookieConfig: cookieConfig,
	}
}

func (h *AuthHandler) Routes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/user/register", h.Register)
		r.Post("/user/login", h.Login)
		r.Post("/user/logout", h.Logout)
		r.Get("/user", h.WhoIAm)
	})
}
