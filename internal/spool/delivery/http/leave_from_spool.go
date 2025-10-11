package deliveryHTTP

import (
	"encoding/json"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) LeaveFromSpool(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", lib.StatusUnauthorized)
		return
	}

	var req dto.LeaveFromSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.usecase.LeaveFromSpool(r.Context(), usecase.LeaveFromSpoolInput{
		UserID:  userID,
		SpoolID: req.SpoolID,
	})
	if err != nil {
		h.logger.Error("failed to leave spool", zap.Error(err))
		http.Error(w, "failed to leave spool", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(lib.StatusOK)
}
