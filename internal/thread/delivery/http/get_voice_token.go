package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetVoiceToken(w http.ResponseWriter, r *http.Request) {
	threadIDstr := chi.URLParam(r, "thread_id")
	id, err := strconv.ParseUint(threadIDstr, 10, 32)
	if err != nil {
		lib.WriteError(w, "parameter thread_id must be a valid integer", lib.StatusUnauthorized)
		return
	}
	threadID := uint(id)

	ctx := r.Context()

	username, err := auth.GetUsernameFromContext(ctx)
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.GetVoiceTokenInput{
		UserID:   userID,
		Username: username,
		ThreadID: threadID,
	}

	tokenOutput, err := h.roomUsecase.GetVoiceToken(ctx, input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to generate voice token", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.GetVoiceTokenResponse{
		Token:    tokenOutput.Token,
		TurnURLs: tokenOutput.TurnURLs,
		TurnUser: tokenOutput.TurnUser,
		TurnPass: tokenOutput.TurnPass,
		TurnTTL:  tokenOutput.TurnTTL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode voice token response", zap.Error(err))
	}
}
