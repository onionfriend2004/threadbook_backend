package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) InviteMemberInSpool(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	username, err := auth.GetUsernameFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.InviteMemberInSpoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.usecase.InviteMemberInSpool(r.Context(), usecase.InviteMemberInSpoolInput{
		UserID:          userID,
		Username:        username,
		SpoolID:         req.SpoolID,
		MemberUsernames: req.MemberUsernames,
	})
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to invite members", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.WriteHeader(lib.StatusOK)
}
