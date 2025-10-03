package deliveryHTTP

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "not authenticated", lib.StatusUnauthorized)
			return
		}
		http.Error(w, "bad request", lib.StatusBadRequest)
		return
	}

	if err := h.usecase.SignOutUser(r.Context(), cookie.Value); err != nil {
		switch {
		case err == usecase.ErrInvalidInput:
			http.Error(w, "invalid session", lib.StatusBadRequest)
		default:
			h.logger.Error("failed to sign out user", zap.Error(err))
			http.Error(w, "internal server error", lib.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, h.cookieConfig.ToHTTPCookie("", -1))
	w.WriteHeader(lib.StatusNoContent)
}
