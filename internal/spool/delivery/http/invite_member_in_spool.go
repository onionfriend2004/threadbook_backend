package deliveryHTTP

import (
	"errors"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
)

func (h *SpoolHandler) InviteMemberInSpool(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", lib.StatusUnauthorized)
		return
	}

	var req dto.InviteMemberInSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	err = h.usecase.InviteMemberInSpool(r.Context(), usecase.InviteMemberInSpoolInput{
		UserID:          userID,
		SpoolID:         req.SpoolID,
		MemberUsernames: req.MemberUsernames,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			lib.WriteError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			lib.WriteError(w, "spool not found", http.StatusNotFound)
			return
		}
		lib.WriteError(w, "internal error", http.StatusInternalServerError)
	}

	w.WriteHeader(lib.StatusOK)
}
