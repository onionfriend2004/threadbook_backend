package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

// нужен ли?
func (h *SpoolHandler) UpdateSpool(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	spool, err := h.usecase.UpdateSpool(r.Context(), usecase.UpdateSpoolInput{
		SpoolID:    req.SpoolID,
		Name:       req.Name,
		BannerLink: req.BannerLink,
	})
	if err != nil {
		h.logger.Error("failed to update spool", zap.Error(err))
		lib.WriteError(w, "failed to update spool", lib.StatusInternalServerError)
		return
	}

	resp := dto.UpdateSpoolResponse{
		SpoolID:    spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
