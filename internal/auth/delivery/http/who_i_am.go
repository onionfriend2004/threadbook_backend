package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

// And darlin' you're the reason why I am who I am
// This you and I was a surprise, it wasn't part of the plan
// I'll bring you down again just like I do when things get shaky
// I'm sorry for the mood, but I've been dying lately
//
// @ ​iamjakehill - ​dying lately

func (h *AuthHandler) WhoIAm(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		if err == http.ErrNoCookie {
			lib.WriteError(w, "not authenticated", lib.StatusUnauthorized)
			return
		}
		lib.WriteError(w, "bad request", lib.StatusBadRequest)
		return
	}

	user, err := h.usecase.AuthenticateUser(r.Context(), cookie.Value)
	if err != nil {
		switch {
		case err == usecase.ErrSessionNotFound, err == usecase.ErrUserNotFound:
			lib.WriteError(w, "not authenticated", lib.StatusUnauthorized)
		default:
			h.logger.Error("failed to authenticate user", zap.Error(err))
			lib.WriteError(w, "internal server error", lib.StatusInternalServerError)
		}
		return
	}

	resp := dto.AuthenticateResponse{
		Email:    user.Email,
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
