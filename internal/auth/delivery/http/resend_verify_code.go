package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *AuthHandler) ResendVerifyCode(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		if err == http.ErrNoCookie {
			lib.WriteError(w, "not authenticated", lib.StatusUnauthorized)
			return
		}
		lib.WriteError(w, "bad request", lib.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		lib.WriteError(w, "invalid cookie value", lib.StatusUnauthorized)
		return
	}

	if err := h.usecase.ResendVerifyCode(r.Context(), userID); err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)

		// Логируем Warn для 4xx, Error для 5xx
		if code >= 500 {
			h.logger.Error("failed to resend verification code", zap.Error(err))
		} else {
			h.logger.Warn("failed to resend verification code", zap.Error(err))
		}

		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.WriteHeader(lib.StatusNoContent)
}
