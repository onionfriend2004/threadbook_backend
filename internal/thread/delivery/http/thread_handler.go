package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
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

		r.With(validator.ValidateJSONMiddleware(dto.JoinNoAuthSessionRequest{})).
			Post("/join/{no_auth_session}", h.JoinNoAuthSession)

		// r.Use(auth.AuthMiddleware(authenticator))
		r.With(auth.AuthMiddleware(authenticator)).Group(func(r chi.Router) {
			r.Put("/close", h.Close)
			r.Get("/", h.GetBySpoolID)
			r.Get("/ws/token", h.GetSubscribeToken)

			r.Get("/create_link", h.CreateNoAuthSession)
			r.With(validator.ValidateJSONMiddleware(dto.ThreadCreateRequest{})).
				Post("/create", h.Create)

			r.With(validator.ValidateJSONMiddleware(dto.InviteRequest{})).
				Post("/invite", h.InviteToThread)

			r.With(validator.ValidateJSONMiddleware(dto.GetVoiceTokenRequest{})).
				Post("/sfu/token", h.GetVoiceToken)

			r.With(validator.ValidateJSONMiddleware(dto.UpdateThreadRequest{})).
				Put("/update", h.Update)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/messages", h.GetMessages)

				r.With(validator.ValidateJSONMiddleware(dto.SendMessageRequest{})).
					Post("/messages", h.SendMessage)
			})
		})
	})
}
