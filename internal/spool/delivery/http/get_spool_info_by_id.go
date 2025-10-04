package deliveryHTTP

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetSpoolInfoById(w http.ResponseWriter, r *http.Request) {
	spoolID, err := lib.ParseIntParam(r, "spoolID")
	if err != nil {
		lib.WriteError(w, "invalid spool id", lib.StatusBadRequest)
		return
	}

	spool, err := h.usecase.GetSpoolInfoById(r.Context(), usecase.GetSpoolInfoByIdInput{SpoolID: spoolID})
	if err != nil {
		h.logger.Error("failed to get spool info", zap.Error(err))
		lib.WriteError(w, "failed to get spool info", lib.StatusInternalServerError)
		return
	}

	resp := dto.GetSpoolInfoByIdResponse{
		ID:         spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
		CreatedAt:  spool.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  spool.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	lib.WriteJSON(w, resp, lib.StatusOK)
}
