package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.ThreadCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	createdThread, err := h.usecase.Create(r.Context(), req.title, req.spoolID, typeThread)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
