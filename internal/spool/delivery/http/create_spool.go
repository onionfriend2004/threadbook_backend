package deliveryHTTP

import (
	"encoding/json"
	"net/http"

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

	lib.WriteJSON(w, resp, lib.StatusOK)
}
