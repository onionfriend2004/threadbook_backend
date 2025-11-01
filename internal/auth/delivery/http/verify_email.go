package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
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
		code, clientErr := apperrors.GetErrAndCodeToSend(err)

		if code >= 500 {
			h.logger.Error("failed to verify email", zap.Error(err))
		} else {
			h.logger.Warn("failed to verify email", zap.Error(err))
		}

		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.WriteHeader(lib.StatusNoContent)
}
