package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"go.uber.org/zap"
)

func (h *AuthHandler) UpgradeGuestToUser(w http.ResponseWriter, r *http.Request) {
	req := validator.GetValidatedBody[dto.RegisterRequest](r)
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
		code, clientErr := apperrors.GetErrAndCodeToSend(err)

		if code >= 500 {
			h.logger.Error("failed to authenticate user", zap.Error(err))
		} else {
			h.logger.Warn("failed to authenticate user", zap.Error(err))
		}

		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	// fmt.Printf("User ID: %d\n", user.ID)
	// gRPC post online status

	user, err = h.usecase.UpgradeGuestToUser(r.Context(), usecase.UpgradeGuestToUserInput{
		UserID:   user.ID,
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to register user", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	session, err := h.usecase.CreateSessionForUser(r.Context(), user)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to create session", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	//TODO: Время сессии хардкоднуто
	http.SetCookie(w, h.cookieConfig.ToHTTPCookie(session.ID, 604800))

	resp := dto.RegisterResponse{
		Email:    user.Email,
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
