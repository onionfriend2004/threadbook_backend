package deliveryHTTP

import (
	"encoding/json"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) UpdateSpool(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	spool, err := h.usecase.UpdateSpool(r.Context(), usecase.UpdateSpoolInput{
		ID:         req.ID,
		Name:       req.Name,
		BannerLink: req.BannerLink,
	})
	if err != nil {
		h.logger.Error("failed to update spool", zap.Error(err))
		lib.WriteError(w, "failed to update spool", lib.StatusInternalServerError)
		return
	}

	resp := dto.UpdateSpoolResponse{
		ID:         spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
	}

	lib.WriteJSON(w, resp, lib.StatusOK)
}
