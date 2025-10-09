package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

type SpoolHandler struct {
	usecase usecase.SpoolUsecaseInterface
	logger  *zap.Logger
}

func NewSpoolHandler(usecase usecase.SpoolUsecaseInterface, logger *zap.Logger) *SpoolHandler {
	return &SpoolHandler{
		usecase: usecase,
		logger:  logger,
	}
}

func (h *SpoolHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/spool", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))
		r.Post("/", h.CreateSpool)
		r.Post("/leave", h.LeaveFromSpool)
		r.Get("/user", h.GetUserSpoolList)
		r.Post("/invite", h.InviteMemberInSpool)
		r.Put("/", h.UpdateSpool)
		r.Get("/{spoolID}", h.GetSpoolInfoById)
		r.Get("/{spoolID}/members", h.GetSpoolMembers)
	})
}
