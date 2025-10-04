package deliveryHTTP

import (
	"encoding/json"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) LeaveFromSpool(w http.ResponseWriter, r *http.Request) {
	var req dto.LeaveFromSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	err := h.usecase.LeaveFromSpool(r.Context(), usecase.LeaveFromSpoolInput{
		UserID:  req.UserID,
		SpoolID: req.SpoolID,
	})
	if err != nil {
		h.logger.Error("failed to leave spool", zap.Error(err))
		lib.WriteError(w, "failed to leave spool", lib.StatusInternalServerError)
		return
	}

	lib.WriteJSON(w, dto.LeaveFromSpoolResponse{Success: true}, lib.StatusOK)
}
