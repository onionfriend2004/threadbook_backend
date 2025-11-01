package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
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

	limit := 50 // default limit
	offset := 0 // default offset

	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	input := usecase.GetMessagesInput{
		ThreadID: threadID,
		Limit:    limit,
		Offset:   offset,
	}

	msgs, err := h.messageUsecase.GetMessages(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to get messages", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := make([]dto.MessageResponse, 0, len(msgs))
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
