package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) CreateSpool(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	spool, err := h.usecase.CreateSpool(r.Context(), usecase.CreateSpoolInput{
		Name:       req.Name,
		BannerLink: req.BannerLink,
	})
	if err != nil {
		h.logger.Error("failed to create spool", zap.Error(err))
		lib.WriteError(w, "failed to create spool", lib.StatusInternalServerError)
		return
	}

	resp := dto.CreateSpoolResponse{
		ID:         spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
