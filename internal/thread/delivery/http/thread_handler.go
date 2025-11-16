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
	logger         *zap.Logger
}

func NewThreadHandler(
	threadUC usecase.ThreadUsecaseInterface,
	messageUC *usecase.MessageUsecase,
	roomUC usecase.RoomUsecaseInterface,
	cookieConfig *config.CookieConfig,
	logger *zap.Logger,
) *ThreadHandler {
	return &ThreadHandler{
		threadUsecase:  threadUC,
		messageUsecase: messageUC,
		roomUsecase:    roomUC,
		cookieConfig:   cookieConfig,
		logger:         logger,
	}
}

func (h *ThreadHandler) Routes(r chi.Router, authenticator auth.AuthenticatorInterface) {
	r.Route("/thread", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authenticator))
		r.With(auth.GuestMiddleware(authenticator)).Group(func(r chi.Router) {
			// Управление тредом
			r.Get("/", h.GetBySpoolID)
			r.With(validator.ValidateJSONMiddleware(dto.ThreadCreateRequest{})).
				Post("/", h.Create)

			r.Route("/{thread_id}", func(r chi.Router) {
				r.Get("/", h.GetThreadUsers)

				// Управление тредом
				r.Put("/close", h.Close)
				r.With(validator.ValidateJSONMiddleware(dto.UpdateThreadRequest{})).
					Put("/update", h.Update)
				r.With(validator.ValidateJSONMiddleware(dto.InviteRequest{})).
					Post("/invite", h.InviteToThread)
				// Инвайт линки
				r.Delete("/invite-link", h.DeleteInviteLink)
				r.Get("/invite-link", h.GetThreadInviteLinks)
				r.With(validator.ValidateJSONMiddleware(dto.CreateInviteLinkRequest{})).
					Post("/invite-link/create", h.CreateInviteLink)

				r.Get("/sfu/token", h.GetVoiceToken)

				r.Get("/messages", h.GetMessages)
				r.With(validator.ValidateJSONMiddleware(dto.SendMessageRequest{})).
					Post("/messages", h.SendMessage)
			})

			r.Get("/invite-link/join/{invite_token}", h.JoinToThread)
			r.Get("/ws/token", h.GetSubscribeToken)
		})
	})
}
