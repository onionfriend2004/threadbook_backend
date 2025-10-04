package deliveryHTTP

import (
	"net/http"

	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"go.uber.org/zap"
)

func (h *ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.ThreadCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return

	}
	spoolID, err := strconv.Atoi(req.SpoolID)
	if err != nil {
		h.logger.Warn("failed string to int spool_id", zap.Error(err))
		return
	}
	createdThread, err := h.usecase.CreateThread(r.Context(), req.Title, spoolID, req.TypeThread)
	if err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}

	resp := dto.ThreadCreateResponse{
		ID:        createdThread.ID,
		SpoolID:   createdThread.SpoolID,
		Title:     createdThread.Title,
		Type:      createdThread.Type,
		IsClosed:  createdThread.IsClosed,
		CreatedAt: createdThread.CreatedAt,
		UpdatedAt: createdThread.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
