package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetSpoolInviteLinks(w http.ResponseWriter, r *http.Request) {
	spoolIDstr := chi.URLParam(r, "spool_id")
	spoolIDInt, err := strconv.Atoi(spoolIDstr)
	if err != nil || spoolIDInt <= 0 {
		lib.WriteError(w, "invalid thread_id", http.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.GetSpoolInviteLinksInput{
		UserID:  userID,
		SpoolID: spoolID,
	}

	links, err := h.usecase.GetSpoolInviteLinks(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to get thread links", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}
	resp := dto.GetSpoolInviteLinksResponse{
		InviteLinks: links,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
