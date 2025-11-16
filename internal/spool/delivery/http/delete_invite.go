package deliveryHTTP

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) DeleteInviteLink(w http.ResponseWriter, r *http.Request) {
	threadLink := r.URL.Query().Get("spool_link")
	if threadLink == "" {
		lib.WriteError(w, "missing thread_link", lib.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.DeleteInviteLinkInput{
		UserID: userID,
		Link:   threadLink,
	}

	err = h.usecase.DeleteInviteLink(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to delete link", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
}
