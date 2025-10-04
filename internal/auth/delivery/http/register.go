package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	user, err := h.usecase.SignUpUser(r.Context(), usecase.SignUpInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case err == usecase.ErrUserAlreadyExists:
			lib.WriteError(w, "user already exists", lib.StatusConflict)
		case err == usecase.ErrInvalidInput:
			lib.WriteError(w, "invalid input", lib.StatusBadRequest)
		default:
			h.logger.Error("failed to register user", zap.Error(err))
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

	resp := dto.RegisterResponse{
		Email:    user.Email,
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
