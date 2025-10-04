package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

type VerifyEmailRequest struct {
	Code int `json:"code"`
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID := 1 // userID, err := xxx.GetUserIDFromContext(r.Context())
	/*if err != nil {
		lib.WriteError(w, "user not authenticated", lib.StatusUnauthorized)
		return
	}*/

	var req VerifyEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid request body", lib.StatusBadRequest)
		return
	}

	if req.Code < 100000 || req.Code > 999999 {
		lib.WriteError(w, "code must be 6 digits", http.StatusBadRequest)
		return
	}

	if err := h.usecase.VerifyUserEmail(r.Context(), userID, req.Code); err != nil {
		switch err {
		case usecase.ErrCodeIncorrect:
			lib.WriteError(w, "invalid verification code", http.StatusBadRequest)
		case usecase.ErrInvalidInput:
			lib.WriteError(w, "invalid input", http.StatusBadRequest)
		default:
			h.logger.Error("failed to verify email", zap.Error(err))
			lib.WriteError(w, "failed to verify email", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(lib.StatusNoContent)
}
