package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	req := validator.GetValidatedBody[dto.ThreadCreateRequest](r)
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", lib.StatusUnauthorized)
		return
	}

	input := usecase.CreateThreadInput{
		Title:       req.Title,
		SpoolID:     req.SpoolID,
		OwnerID:     userID,
		ThreadType:  req.ThreadType,
		AccessLevel: req.AccessLevel,
	}

	createdThread, err := h.threadUsecase.CreateThread(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to create thread", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.ThreadCreateResponse{
		ID: createdThread.ID,
		// SpoolID:     createdThread.SpoolID,
		AccessLevel: createdThread.AccessLevel,
		Title:       createdThread.Title,
		Type:        createdThread.Type,
		IsClosed:    createdThread.IsClosed,
		CreatedAt:   createdThread.CreatedAt,
		UpdatedAt:   createdThread.UpdatedAt,
	}
	if createdThread.SpoolID != nil {
		resp.SpoolID = *createdThread.SpoolID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
