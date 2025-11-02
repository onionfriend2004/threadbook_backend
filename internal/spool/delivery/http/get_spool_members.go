package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetSpoolMembers(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	spoolIDStr := chi.URLParam(r, "spoolID")
	spoolIDInt, err := strconv.Atoi(spoolIDStr)
	if err != nil || spoolIDInt <= 0 {
		lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)

	users, err := h.usecase.GetSpoolMembers(r.Context(), usecase.GetSpoolMembersInput{
		UserID:  userID,
		SpoolID: spoolID,
	})
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to get spool members", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	resp := dto.GetSpoolMembersResponse{}
	for _, u := range users {
		resp.Members = append(resp.Members, dto.MemberShortInfo{
			Username: u.Username,
			// Avatar:   u.AvatarLink,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
