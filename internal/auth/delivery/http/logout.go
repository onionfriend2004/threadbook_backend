package deliveryHTTP

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		if err == http.ErrNoCookie {
			lib.WriteError(w, "not authenticated", lib.StatusUnauthorized)
			return
		}
		lib.WriteError(w, "bad request", lib.StatusBadRequest)
		return
	}

	if err := h.usecase.SignOutUser(r.Context(), cookie.Value); err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)

		// Логируем Warn для 4xx, Error для 5xx
		if code >= 500 {
			h.logger.Error("failed to sign out user", zap.Error(err))
		} else {
			h.logger.Warn("failed to sign out user", zap.Error(err))
		}

		lib.WriteError(w, clientErr.Error(), code)
		return
	}
	http.SetCookie(w, h.cookieConfig.ToHTTPCookie("", -1))
	w.WriteHeader(lib.StatusNoContent)
}
