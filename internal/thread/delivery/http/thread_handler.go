package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

type ThreadHandler struct {
	threadUsecase  usecase.ThreadUsecaseInterface
	messageUsecase *usecase.MessageUsecase
	roomUsecase    usecase.RoomUsecaseInterface
	cookieConfig   *config.CookieConfig
	fileConfig     *config.FileConfig
	logger         *zap.Logger
}

func NewThreadHandler(
	threadUC usecase.ThreadUsecaseInterface,
	messageUC *usecase.MessageUsecase,
	roomUC usecase.RoomUsecaseInterface,
	cookieConfig *config.CookieConfig,
	fileConfig *config.FileConfig,
	logger *zap.Logger,
) *ThreadHandler {
	return &ThreadHandler{
		threadUsecase:  threadUC,
		messageUsecase: messageUC,
		roomUsecase:    roomUC,
		cookieConfig:   cookieConfig,
		fileConfig:     fileConfig,
		logger:         logger,
	}
}

func (h *ThreadHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/thread", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))
		r.With(auth.GuestMiddleware(authenticator)).Group(func(r chi.Router) {
			r.Put("/close", h.Close)

			r.With(validator.ValidateJSONMiddleware(dto.UpdateThreadRequest{})).
				Put("/update", h.Update)

			r.With(validator.ValidateJSONMiddleware(dto.ThreadCreateRequest{})).
				Post("/create", h.Create)

			r.Get("/invite-link/create", h.CreateInviteLink)
			r.Delete("/invite-link", h.DeleteInviteLink)
			r.Get("/invite-link/{thread_id}", h.GetThreadInviteLinks)

			r.With(validator.ValidateJSONMiddleware(dto.InviteRequest{})).
				Post("/invite", h.InviteToThread)

		})

		r.Get("/ws/token", h.GetSubscribeToken)
		r.Get("/", h.GetBySpoolID)
		r.Get("/invite-link/join/{thread_link}", h.JoinToThread)

		r.With(validator.ValidateJSONMiddleware(dto.GetVoiceTokenRequest{})).
			Post("/sfu/token", h.GetVoiceToken)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/messages", h.GetMessages)

			r.Post("/messages", h.SendMessage)

			r.With(validator.ValidateJSONMiddleware(dto.UpdateMessageRequest{})).
				Put("/messages/{message_id}", h.UpdateMessage)

			r.Delete("/messages/{message_id}", h.DeleteMessage)
		})

	})
}
