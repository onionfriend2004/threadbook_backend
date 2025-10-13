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

// GetVoiceToken выдаёт токен для подключения к голосовой комнате треда
func (h *ThreadHandler) GetVoiceToken(w http.ResponseWriter, r *http.Request) {
	var req dto.GetVoiceTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	username, err := auth.GetUsernameFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := h.usecase.GetVoiceToken(r.Context(), username, req.ThreadID)
	if err != nil {
		switch {
		case err == usecase.ErrInvalidInput:
			lib.WriteError(w, "invalid input", lib.StatusBadRequest)
		case err == usecase.ErrThreadNotFound:
			lib.WriteError(w, "thread not found", lib.StatusNotFound)
		case err == usecase.ErrFaildToEnsureRoom:
			lib.WriteError(w, "failed to prepare voice room", lib.StatusInternalServerError)
		default:
			h.logger.Error("failed to generate voice token", zap.Error(err))
			lib.WriteError(w, "internal server error", lib.StatusInternalServerError)
		}
		return
	}

	resp := dto.GetVoiceTokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode voice token response", zap.Error(err))
		return
	}
}
