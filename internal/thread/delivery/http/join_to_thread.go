package deliveryHTTP

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) JoinToThread(w http.ResponseWriter, r *http.Request) {
	link := chi.URLParam(r, "thread_link")
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	input := usecase.JoinToThreadInput{
		Link:   link,
		UserID: userID,
	}
	if err := h.threadUsecase.JoinToThread(r.Context(), input); err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to invite user", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
