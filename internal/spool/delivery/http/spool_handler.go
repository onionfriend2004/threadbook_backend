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
		r.With(auth.GuestMiddleware(authenticator)).Group(func(r chi.Router) {
			r.Post("/", h.CreateSpool)
			r.Put("/", h.UpdateSpool)

			r.With(validator.ValidateJSONMiddleware(dto.LeaveFromSpoolRequest{})).
				Post("/leave", h.LeaveFromSpool)

			r.With(validator.ValidateJSONMiddleware(dto.InviteMemberInSpoolRequest{})).
				Post("/invite", h.InviteMemberInSpool)
		})
		r.Get("/user", h.GetUserSpoolList)

		r.Route("/{spoolID}", func(r chi.Router) {
			r.Get("/", h.GetSpoolInfoById)
			r.Route("/members", func(r chi.Router) {
				r.Get("/", h.GetSpoolMembers)
				r.With(validator.ValidateJSONMiddleware(dto.AccessLevelRequest{})).
					Post("/{username}/access-level", h.AccessLevel)
			})
			r.With(validator.ValidateJSONMiddleware(dto.CreateInviteLinkRequest{})).
				Post("/invite-link/create", h.CreateInviteLink)
			r.Delete("/invite-link", h.DeleteInviteLink)
			r.Post("/invite-link", h.GetSpoolInviteLinks)
		})

		r.Get("/invite-link/join/{spool_link}", h.JoinToSpool)
	})
}
