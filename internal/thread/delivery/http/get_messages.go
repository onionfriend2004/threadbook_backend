package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	threadID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid thread id", lib.StatusBadRequest)
		return
	}
	threadID := uint(threadID64)

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	input := usecase.GetMessagesInput{
		ThreadID: threadID,
		Limit:    limit,
		Offset:   offset,
	}

	msgs, err := h.messageUsecase.GetMessages(r.Context(), input)
	if err != nil {
		h.logger.Warn("failed to get messages", zap.Error(err))
		lib.WriteError(w, "failed to get messages", lib.StatusInternalServerError)
		return
	}
	var resp []dto.MessageResponse
	for _, m := range msgs {
		resp = append(resp, dto.MessageResponse{
			ID:        m.ID,
			ThreadID:  m.ThreadID,
			Username:  m.User.Username,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
