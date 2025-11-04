package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetThreadInviteLinks(w http.ResponseWriter, r *http.Request) {
	threadIDstr := chi.URLParam(r, "thread_id")
	threadIDInt, err := strconv.Atoi(threadIDstr)
	if err != nil || threadIDInt <= 0 {
		lib.WriteError(w, "invalid thread_id", http.StatusBadRequest)
		return
	}
	threadID := uint(threadIDInt)

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.GetThreadInviteLinksInput{
		UserID:   userID,
		ThreadID: threadID,
	}

	links, err := h.threadUsecase.GetThreadInviteLinks(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to get thread links", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}
	resp := dto.GetThreadInviteLinksResponse{
		InviteLinks: links,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
