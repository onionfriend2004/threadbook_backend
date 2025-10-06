package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
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

func (h *ThreadHandler) Routes(r chi.Router) {
	r.Route("/thread", func(r chi.Router) {
		r.Post("/create", h.Create)
		r.Get("/close", h.Close)
		r.Get("/", h.GetBySpoolID)

		r.Post("/sfu/token", h.GetVoiceToken)
	})
}
