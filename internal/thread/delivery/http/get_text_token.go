package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetSubscribeToken(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := h.messageUsecase.GetSubscribeToken(r.Context(), usecase.GetSubscribeTokenInput{
		UserID: userID,
	})
	if err != nil {
		h.logger.Warn("failed to generate subscribe token", zap.Error(err))
		lib.WriteError(w, "failed to generate token", lib.StatusInternalServerError)
		return
	}

	resp := dto.SubscribeTokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
