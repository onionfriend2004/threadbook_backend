package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) InviteMemberInSpool(w http.ResponseWriter, r *http.Request) {
	var req dto.InviteMemberInSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	err := h.usecase.InviteMemberInSpool(r.Context(), usecase.InviteMemberInSpoolInput{
		SpoolID:  req.SpoolID,
		MemberID: req.MemberID,
	})
	if err != nil {
		h.logger.Error("failed to invite member", zap.Error(err))
		lib.WriteError(w, "failed to invite member", lib.StatusInternalServerError)
		return
	}

	resp := dto.InviteMemberInSpoolResponse{Success: true}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
