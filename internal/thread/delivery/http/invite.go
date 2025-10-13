package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"go.uber.org/zap"
)

func (h *ThreadHandler) InviteToThread(w http.ResponseWriter, r *http.Request) {
	var req dto.InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode invite request", zap.Error(err))
		lib.WriteError(w, "invalid request body", lib.StatusBadRequest)
		return
	}

	if req.ThreadID == 0 || req.InviteeID == 0 {
		lib.WriteError(w, "thread_id and invitee_id are required", lib.StatusBadRequest)
		return
	}

	inviterID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.usecase.InviteToThread(r.Context(), int(inviterID), req.InviteeID, req.ThreadID)
	if err != nil {
		lib.WriteError(w, "failed to invite user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
