package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetSubscribeToken(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	spoolIDParam := r.URL.Query().Get("spool_id")

	var result usecase.ConnectAndSubscribeTokens

	if spoolIDParam == "" {
		result, err = h.messageUsecase.GetUserOnlyTokens(r.Context(), userID)
	} else {
		spoolID64, errConv := strconv.ParseUint(spoolIDParam, 10, 64)
		if errConv != nil {
			lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
			return
		}
		result, err = h.messageUsecase.GetTokensBySpool(r.Context(), userID, uint(spoolID64))
	}

	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to generate subscribe tokens", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Warn("failed to encode subscribe token response", zap.Error(err))
	}
}
