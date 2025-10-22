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

func (h *ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.ThreadCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid JSON", lib.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.CreateThreadInput{
		Title:      req.Title,
		SpoolID:    req.SpoolID,
		OwnerID:    int(userID),
		TypeThread: req.TypeThread,
	}

	createdThread, err := h.usecase.CreateThread(r.Context(), input)
	if err != nil {
		h.logger.Warn("failed to create thread", zap.Error(err))
		lib.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := dto.ThreadCreateResponse{
		ID:        createdThread.ID,
		SpoolID:   createdThread.SpoolID,
		Title:     createdThread.Title,
		Type:      createdThread.Type,
		IsClosed:  createdThread.IsClosed,
		CreatedAt: createdThread.CreatedAt,
		UpdatedAt: createdThread.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
