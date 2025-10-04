package deliveryHTTP

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetUserSpoolList(w http.ResponseWriter, r *http.Request) {
	userID, err := lib.ParseIntParam(r, "userID")
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

	lib.WriteJSON(w, resp, lib.StatusOK)
}
