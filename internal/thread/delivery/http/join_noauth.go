package deliveryHTTP

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) JoinNoAuthSession(w http.ResponseWriter, r *http.Request) {
	req := validator.GetValidatedBody[dto.JoinNoAuthSessionRequest](r)
	ctx := r.Context()
	session := chi.URLParam(r, "no_auth_session")
	input := usecase.GetNoAuthVoiceTokenInput{
		Nickname:      req.Nickname,
		NoAuthSession: session,
	}

	token, err := h.roomUsecase.GetNoAuthVoiceToken(ctx, input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to generate voice token", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.GetVoiceTokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode voice token response", zap.Error(err))
	}
}
