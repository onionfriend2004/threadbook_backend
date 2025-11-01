package deliveryHTTP

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/profile/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/profile/usecase"
	"go.uber.org/zap"
)

func (h *ProfileHandler) EditProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.fileConfig.GetMaxSize("common")); err != nil {
		lib.WriteError(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	// Достаём userID из сессии
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	username, err := auth.GetUsernameFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	nickname := strings.TrimSpace(r.FormValue("nickname"))
	if len(nickname) > 32 {
		lib.WriteError(w, "nickname too long (max 32 chars)", http.StatusBadRequest)
		return
	}

	var avatarInput *usecase.Avatar
	file, fileHeader, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		if !h.fileConfig.ValidateSize("profile_avatar", fileHeader.Size) {
			maxSizeMB := h.fileConfig.GetMaxSize("profile_avatar") >> 20
			lib.WriteError(w, fmt.Sprintf("avatar size exceeds limit of %dMB", maxSizeMB), http.StatusBadRequest)
			return
		}

		if !h.fileConfig.IsAllowedFormat(fileHeader.Filename) {
			allowedFormats := strings.Join(h.fileConfig.GetAllowedFormats(), ", ")
			lib.WriteError(w, fmt.Sprintf("allowed formats: %s", allowedFormats), http.StatusBadRequest)
			return
		}

		avatarInput = &usecase.Avatar{
			File:        file,
			Size:        fileHeader.Size,
			Filename:    fileHeader.Filename,
			Filetype:    "avatar",
			ContentType: h.fileConfig.GetContentTypeByExtension(fileHeader.Filename),
		}
	} else if err != http.ErrMissingFile {
		lib.WriteError(w, "invalid avatar file", http.StatusBadRequest)
		return
	}

	profile, err := h.usecase.UpdateProfile(r.Context(), usecase.UpdateProfileInput{
		UserID:   int(userID),
		Nickname: &nickname,
		Avatar:   avatarInput,
	})
	if err != nil {
		h.logger.Error("failed to update profile", zap.Error(err))
		lib.WriteError(w, "failed to update profile", http.StatusInternalServerError)
		return
	}
	resp := dto.UpdateProfileResponse{
		Username:   username,
		Nickname:   profile.Nickname,
		AvatarLink: profile.AvatarLink,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
