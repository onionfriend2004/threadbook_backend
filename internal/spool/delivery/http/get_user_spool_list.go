package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetUserSpoolList(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		lib.WriteError(w, "invalid user id", lib.StatusBadRequest)
		return
	}

	spools, err := h.usecase.GetUserSpoolList(r.Context(), usecase.GetUserSpoolListInput{UserID: userID})
	if err != nil {
		h.logger.Error("failed to get user spools", zap.Error(err))
		lib.WriteError(w, "failed to get spools", lib.StatusInternalServerError)
		return
	}

	resp := dto.GetUserSpoolListResponse{}
	for _, s := range spools {
		resp.Spools = append(resp.Spools, dto.SpoolShortInfo{
			ID:         s.ID,
			Name:       s.Name,
			BannerLink: s.BannerLink,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
