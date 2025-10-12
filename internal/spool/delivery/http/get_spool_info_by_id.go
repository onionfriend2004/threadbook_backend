package deliveryHTTP

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetSpoolInfoById(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	spoolIDStr := chi.URLParam(r, "spoolID")
	spoolIDInt, err := strconv.Atoi(spoolIDStr)
	if err != nil || spoolIDInt < 0 {
		lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)

	spool, err := h.usecase.GetSpoolInfoById(r.Context(), usecase.GetSpoolInfoByIdInput{UserID: userID, SpoolID: spoolID})
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			lib.WriteError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			lib.WriteError(w, "spool not found", http.StatusNotFound)
			return
		}
		lib.WriteError(w, "internal error", http.StatusInternalServerError)
	}

	resp := dto.GetSpoolInfoByIdResponse{
		SpoolID:    spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
		CreatedAt:  spool.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  spool.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
