package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
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
		r.Get("/user", h.WhoIAm)

		r.With(validator.ValidateJSONMiddleware(dto.RegisterRequest{})).
			Post("/user/register", h.Register)

		r.With(validator.ValidateJSONMiddleware(dto.LoginRequest{})).
			Post("/user/login", h.Login)

		r.Post("/user/logout", h.Logout)

		r.Post("/email/verify", h.VerifyEmail)

		r.Post("/email/resend", h.ResendVerifyCode)
	})
}
