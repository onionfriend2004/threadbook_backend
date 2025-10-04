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
		SpoolID:         req.SpoolID,
		MemberUsernames: req.MemberUsernames,
	})
	if err != nil {
		h.logger.Error("failed to invite members", zap.Error(err))
		lib.WriteError(w, "failed to invite members", lib.StatusInternalServerError)
		return
	}

	w.WriteHeader(lib.StatusOK)
}
