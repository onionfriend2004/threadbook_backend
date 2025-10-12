package deliveryHTTP

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) CreateSpool(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.fileConfig.GetMaxSize("common")); err != nil {
		lib.WriteError(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем name из form data
	spoolName := strings.TrimSpace(r.FormValue("name"))
	if spoolName == "" {
		lib.WriteError(w, "spool name is required", http.StatusBadRequest)
		return
	}

	var bannerInput *usecase.BannerInput
	file, fileHeader, err := r.FormFile("banner")
	if err == nil {
		defer file.Close()

		if !h.fileConfig.ValidateSize("spool_banner", fileHeader.Size) {
			maxSizeMB := h.fileConfig.GetMaxSize("spool_banner") >> 20
			lib.WriteError(w, fmt.Sprintf("banner size exceeds limit of %dMB", maxSizeMB), http.StatusBadRequest)
			return
		}

		if !h.fileConfig.IsAllowedFormat(fileHeader.Filename) {
			allowedFormats := strings.Join(h.fileConfig.GetAllowedFormats(), ", ")
			lib.WriteError(w, fmt.Sprintf("allowed formats: %s", allowedFormats), http.StatusBadRequest)
			return
		}

		bannerInput = &usecase.BannerInput{
			File:        file,
			Size:        fileHeader.Size,
			Filename:    fileHeader.Filename,
			ContentType: h.fileConfig.GetContentTypeByExtension(fileHeader.Filename),
		}
	} else if err != http.ErrMissingFile {
		lib.WriteError(w, "invalid banner file", http.StatusBadRequest)
		return
	}

	spool, err := h.usecase.CreateSpool(r.Context(), usecase.CreateSpoolInput{
		OwnerID:     userID,
		Name:        spoolName,
		BannerInput: bannerInput,
	})
	if err != nil {
		h.logger.Error("failed to create spool", zap.Error(err))
		lib.WriteError(w, "failed to create spool", http.StatusInternalServerError)
		return
	}

	resp := dto.CreateSpoolResponse{
		SpoolID:    spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
