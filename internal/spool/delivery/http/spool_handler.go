package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

type SpoolHandler struct {
	usecase    usecase.SpoolUsecaseInterface
	logger     *zap.Logger
	fileConfig *config.FileConfig
}

func NewSpoolHandler(u usecase.SpoolUsecaseInterface, logger *zap.Logger, fileConfig *config.FileConfig) *SpoolHandler {
	return &SpoolHandler{
		usecase:    u,
		logger:     logger,
		fileConfig: fileConfig,
	}
}

func (h *SpoolHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/spool", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))
		r.Get("/{spoolID}", h.GetSpoolInfoById)
		r.Get("/{spoolID}/members", h.GetSpoolMembers)
		r.Get("/user", h.GetUserSpoolList)

		r.Post("/", h.CreateSpool)

		r.With(validator.ValidateJSONMiddleware(dto.LeaveFromSpoolRequest{})).
			Post("/leave", h.LeaveFromSpool)

		r.With(validator.ValidateJSONMiddleware(dto.InviteMemberInSpoolRequest{})).
			Post("/invite", h.InviteMemberInSpool)

		r.Put("/", h.UpdateSpool)
	})
}
