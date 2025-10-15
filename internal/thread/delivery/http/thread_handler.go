package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

type ThreadHandler struct {
	usecase usecase.ThreadUsecaseInterface
	logger  *zap.Logger
}

func NewThreadHandler(
	usecase usecase.ThreadUsecaseInterface,
	logger *zap.Logger,
) *ThreadHandler {
	return &ThreadHandler{
		usecase: usecase,
		logger:  logger,
	}
}

func (h *ThreadHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/thread", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))

		r.Post("/create", h.Create)
		r.Put("/close", h.Close)
		r.Get("/", h.GetBySpoolID)
		r.Post("/invite", h.InviteToThread)
		r.Post("/sfu/token", h.GetVoiceToken)
		r.Put("/update", h.Update)
	})
}
