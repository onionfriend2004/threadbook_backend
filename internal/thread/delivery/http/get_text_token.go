package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetSubscribeToken(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tokens, err := h.messageUsecase.GetConnectAndSubscribeTokens(r.Context(), userID)
	if err != nil {
		h.logger.Warn("failed to generate tokens", zap.Error(err))
		lib.WriteError(w, "failed to generate tokens", lib.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.logger.Warn("failed to encode subscribe token response", zap.Error(err))
		return
	}
}
