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

func (h *SpoolHandler) GetSpoolMembers(w http.ResponseWriter, r *http.Request) {
	spoolIDStr := chi.URLParam(r, "spoolID")
	spoolID, err := strconv.Atoi(spoolIDStr)
	if err != nil {
		lib.WriteError(w, "invalid spool id", lib.StatusBadRequest)
		return
	}

	users, err := h.usecase.GetSpoolMembers(r.Context(), usecase.GetSpoolMembersInput{SpoolID: spoolID})
	if err != nil {
		h.logger.Error("failed to get spool members", zap.Error(err))
		lib.WriteError(w, "failed to get spool members", lib.StatusInternalServerError)
		return
	}

	resp := dto.GetSpoolMembersResponse{}
	for _, u := range users {
		resp.Members = append(resp.Members, dto.MemberShortInfo{
			ID:       u.ID,
			Username: u.Username,
			// Nickname: u.Nickname,
			// Avatar:   u.AvatarLink,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		return
	}
}
