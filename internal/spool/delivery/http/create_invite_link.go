package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	threadIDStr := r.URL.Query().Get("thread_id")
	if threadIDStr == "" {
		lib.WriteError(w, "missing thread_id", lib.StatusBadRequest)
		return
	}
	req := validator.GetValidatedBody[dto.CreateInviteLinkRequest](r)

	threadIDint, err := strconv.Atoi(threadIDStr)
	if err != nil || threadIDint < 0 {
		h.logger.Warn("failed string to int thread_id", zap.Error(err))
		lib.WriteError(w, "invalid thread_id", lib.StatusBadRequest)
		return
	}
	threadID := uint(threadIDint)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.CreateInviteLinkInput{
		UserID:    userID,
		SpoolID:   threadID,
		MaxUses:   req.MaxUses,
		ExpiresAt: req.ExpiresAt,
	}

	session, err := h.usecase.CreateInviteLink(r.Context(), input)
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
