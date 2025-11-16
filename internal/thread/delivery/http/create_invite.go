package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	threadIDstr := chi.URLParam(r, "thread_id")
	id, err := strconv.ParseUint(threadIDstr, 10, 32)
	if err != nil {
		lib.WriteError(w, "parameter thread_id must be a valid integer", lib.StatusUnauthorized)
		return
	}
	threadID := uint(id)

	req := validator.GetValidatedBody[dto.CreateInviteLinkRequest](r)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.CreateInviteLinkInput{
		UserID:    userID,
		ThreadID:  threadID,
		MaxUses:   req.MaxUses,
		ExpiresAt: req.ExpiresAt,
	}

	session, err := h.threadUsecase.CreateInviteLink(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to create session", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(session); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
