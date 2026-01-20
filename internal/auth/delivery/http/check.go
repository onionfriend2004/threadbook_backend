package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"go.uber.org/zap"
)

func (h *AuthHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	req := validator.GetValidatedBody[dto.CheckUsernameRequest](r)

	isExist, err := h.usecase.CheckUsername(r.Context(), req.Username)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to register user", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.IsExist{
		IsExist: isExist,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}

func (h *AuthHandler) CheckEmail(w http.ResponseWriter, r *http.Request) {
	req := validator.GetValidatedBody[dto.CheckEmailRequest](r)

	isExist, isValid, err := h.usecase.CheckEmail(r.Context(), req.Email)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to register user", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.IsExistWithValidFlag{
		IsExist:      dto.IsExist{IsExist: isExist},
		IsValidEmail: isValid,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
