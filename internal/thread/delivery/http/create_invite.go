package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	threadIDStr := r.URL.Query().Get("thread_id")
	if threadIDStr == "" {
		lib.WriteError(w, "missing thread_id", lib.StatusBadRequest)
		return
	}

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
		UserID:   userID,
		ThreadID: threadID,
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
