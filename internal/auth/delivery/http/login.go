package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	user, err := h.usecase.SignInUser(r.Context(), usecase.SignInInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case err == usecase.ErrInvalidCredentials:
			lib.WriteError(w, "invalid credentials", lib.StatusUnauthorized)
		case err == usecase.ErrInvalidInput:
			lib.WriteError(w, "invalid input", lib.StatusBadRequest)
		default:
			h.logger.Error("failed to sign in user", zap.Error(err))
			lib.WriteError(w, "internal server error", lib.StatusInternalServerError)
		}
		return
	}

	session, err := h.usecase.CreateSessionForUser(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to create session", zap.Error(err))
		lib.WriteError(w, "internal server error", lib.StatusInternalServerError)
		return
	}

	http.SetCookie(w, h.cookieConfig.ToHTTPCookie(session.ID, 0))

	resp := dto.LoginResponse{
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
