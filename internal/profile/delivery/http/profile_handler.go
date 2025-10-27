package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/profile/usecase"
	"go.uber.org/zap"
)

type ProfileHandler struct {
	usecase    usecase.ProfileUsecaseInterface
	logger     *zap.Logger
	fileConfig *config.FileConfig
}

func NewProfileHandler(
	u usecase.ProfileUsecaseInterface,
	logger *zap.Logger,
	fileConfig *config.FileConfig,
) *ProfileHandler {
	return &ProfileHandler{
		usecase:    u,
		logger:     logger,
		fileConfig: fileConfig,
	}
}

func (h *ProfileHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/profile", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))

		r.Post("/edit", h.EditProfile)
		r.Post("/get", h.GetProfiles)
	})
}
