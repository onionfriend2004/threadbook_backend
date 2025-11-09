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

func (h *ThreadHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	// Получаем thread_id
	threadIDStr := chi.URLParam(r, "id")
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

	// Получаем тело запроса
	req := validator.GetValidatedBody[dto.UpdateMessageRequest](r)

	input := usecase.UpdateMessageInput{
		ThreadID:  threadID,
		MessageID: messageID,
		UserID:    userID,
		Content:   req.Content,
	}

	msg, err := h.messageUsecase.UpdateMessage(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to update message", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.SendMessageResponse{
		Message: dto.MessageResponse{
			ID:        msg.ID,
			ThreadID:  msg.ThreadID,
			Username:  msg.User.Username,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode update message response", zap.Error(err))
	}
}
