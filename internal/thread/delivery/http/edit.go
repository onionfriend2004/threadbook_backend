package deliveryHTTP

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.UpdateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid request body", lib.StatusBadRequest)
		return
	}

	input := usecase.UpdateThreadInput{
		ID:         req.ID,
		EditorID:   int(userID),
		Title:      req.Title,
		ThreadType: req.Type,
	}

	updatedThread, err := h.usecase.UpdateThread(r.Context(), input)
	if err != nil {
		h.logger.Warn("failed to update thread", zap.Error(err))
		lib.WriteError(w, "failed to update thread", lib.StatusInternalServerError)
		return
	}

	resp := dto.ThreadCreateResponse{
		ID:        updatedThread.ID,
		SpoolID:   updatedThread.SpoolID,
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
