package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

type ThreadHandler struct {
	threadUsecase  usecase.ThreadUsecaseInterface
	messageUsecase *usecase.MessageUsecase
	roomUsecase    usecase.RoomUsecaseInterface
	logger         *zap.Logger
}

func NewThreadHandler(
	threadUC usecase.ThreadUsecaseInterface,
	messageUC *usecase.MessageUsecase,
	roomUC usecase.RoomUsecaseInterface,
	logger *zap.Logger,
) *ThreadHandler {
	return &ThreadHandler{
		threadUsecase:  threadUC,
		messageUsecase: messageUC,
		roomUsecase:    roomUC,
		logger:         logger,
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
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/messages", h.GetMessages)
			r.Post("/messages", h.SendMessage)
		})
		r.Get("/ws/token", h.GetSubscribeToken)
	})
}
