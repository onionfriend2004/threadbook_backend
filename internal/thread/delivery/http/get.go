package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetBySpoolID(w http.ResponseWriter, r *http.Request) {
	spoolIDStr := r.URL.Query().Get("spool_id")
	if spoolIDStr == "" {
		lib.WriteError(w, "missing spool_id", lib.StatusBadRequest)
		return
	}

	spoolIDInt, err := strconv.Atoi(spoolIDStr)
	if err != nil || spoolIDInt < 0 {
		h.logger.Warn("failed string to int spool_id", zap.Error(err))
		lib.WriteError(w, "invalid spool_id", lib.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.GetBySpoolIDInput{
		UserID:  userID,
		SpoolID: spoolID,
	}

	threads, err := h.threadUsecase.GetBySpoolID(r.Context(), input)
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
	}
}
