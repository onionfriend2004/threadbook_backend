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

func (h *ThreadHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", lib.StatusUnauthorized)
		return
	}

	req := validator.GetValidatedBody[dto.UpdateThreadRequest](r)

	input := usecase.UpdateThreadInput{
		ID:          req.ID,
		EditorID:    userID,
		Title:       req.Title,
		ThreadType:  req.Type,
		AccessLevel: req.AccessLevel,
	}

	updatedThread, err := h.threadUsecase.UpdateThread(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to update thread", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.ThreadCreateResponse{
		ID:        updatedThread.ID,
		SpoolID:   *updatedThread.SpoolID,
		Title:     updatedThread.Title,
		Type:      updatedThread.Type,
		IsClosed:  updatedThread.IsClosed,
		CreatedAt: updatedThread.CreatedAt,
		UpdatedAt: updatedThread.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
