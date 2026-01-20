package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	// Получаем thread_id
	threadIDStr := chi.URLParam(r, "thread_id")
	threadID64, err := strconv.ParseUint(threadIDStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid thread id", lib.StatusBadRequest)
		return
	}
	threadID := uint(threadID64)

	// Получаем message_id
	messageIDStr := chi.URLParam(r, "message_id")
	messageID64, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid message id", lib.StatusBadRequest)
		return
	}
	messageID := uint(messageID64)

	// Получаем user_id
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.messageUsecase.DeleteMessage(r.Context(), usecase.DeleteMessageInput{
		ThreadID:  threadID,
		MessageID: messageID,
		UserID:    userID,
	})
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to delete message", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.WriteHeader(lib.StatusNoContent)
}
