package deliveryHTTP

import (
	"net/http"

	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"go.uber.org/zap"
)

func (h *ThreadHandler) Close(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		lib.WriteError(w, "invalid thread id", lib.StatusBadRequest)
		return
	}

	closedThread, err := h.usecase.CloseThread(r.Context(), id)
	if err != nil {
		h.logger.Warn("failed to close thread", zap.Error(err))
		lib.WriteError(w, "failed to close thread", lib.StatusInternalServerError)
		return
	}

	resp := dto.ThreadCreateResponse{
		ID:        closedThread.ID,
		SpoolID:   closedThread.SpoolID,
		Title:     closedThread.Title,
		Type:      closedThread.Type,
		IsClosed:  closedThread.IsClosed,
		CreatedAt: closedThread.CreatedAt,
		UpdatedAt: closedThread.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
