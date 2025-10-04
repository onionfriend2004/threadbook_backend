package deliveryHTTP

import (
	"net/http"

	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetBySpoolID(w http.ResponseWriter, r *http.Request) {
	spoolIDStr := r.URL.Query().Get("spool_id")
	if spoolIDStr == "" {
		lib.WriteError(w, "missing spool_id", lib.StatusBadRequest)
		return
	}

	spoolID, err := strconv.Atoi(spoolIDStr)
	if err != nil {
		h.logger.Warn("failed string to int spool_id", zap.Error(err))
		lib.WriteError(w, "invalid spool_id", lib.StatusBadRequest)
		return
	}

	threads, err := h.usecase.GetBySpoolID(r.Context(), spoolID)
	if err != nil {
		h.logger.Warn("failed to get threads by spool_id", zap.Error(err))
		lib.WriteError(w, "failed to get threads", lib.StatusInternalServerError)
		return
	}

	var resp []dto.ThreadCreateResponse
	for _, t := range threads {
		resp = append(resp, dto.ThreadCreateResponse{
			ID:        t.ID,
			SpoolID:   t.SpoolID,
			Title:     t.Title,
			Type:      t.Type,
			IsClosed:  t.IsClosed,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
