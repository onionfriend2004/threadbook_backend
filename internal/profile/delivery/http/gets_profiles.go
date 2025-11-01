package deliveryHTTP

import (
	"encoding/json"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"

	"github.com/onionfriend2004/threadbook_backend/internal/profile/delivery/dto"
	"go.uber.org/zap"
)

func (h *ProfileHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.GetProfilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Usernames) == 0 {
		lib.WriteError(w, "usernames cannot be empty", http.StatusBadRequest)
		return
	}

	profiles, err := h.usecase.GetProfilesByUsernames(r.Context(), req.Usernames)
	if err != nil {
		h.logger.Error("failed to get profiles", zap.Error(err))
		lib.WriteError(w, "failed to get profiles", http.StatusInternalServerError)
		return
	}

	resp := dto.GetProfilesResponse{
		Profiles: make([]dto.GetProfilesResponseItem, 0, len(profiles)),
	}

	for _, p := range profiles {
		resp.Profiles = append(resp.Profiles, dto.GetProfilesResponseItem{
			Username:   p.Username,
			Nickname:   p.Nickname,
			AvatarLink: p.AvatarLink,
		})
	}

	json.NewEncoder(w).Encode(resp)
}
