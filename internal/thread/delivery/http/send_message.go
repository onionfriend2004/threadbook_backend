package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	threadID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid thread id", lib.StatusBadRequest)
		return
	}
	threadID := uint(threadID64)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid request body", lib.StatusBadRequest)
		return
	}

	input := usecase.SendMessageInput{
		ThreadID: threadID,
		UserID:   userID,
		Content:  req.Content,
	}

	msg, err := h.messageUsecase.SendMessage(r.Context(), input)
	if err != nil {
		h.logger.Warn("failed to send message", zap.Error(err))
		lib.WriteError(w, "failed to send message", lib.StatusInternalServerError)
		return
	}

	resp := dto.SendMessageResponse{
		Message: dto.MessageResponse{
			ID:        msg.ID,
			ThreadID:  msg.ThreadID,
			UserID:    msg.UserID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
