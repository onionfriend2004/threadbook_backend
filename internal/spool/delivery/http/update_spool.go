package deliveryHTTP

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) UpdateSpool(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим multipart/form-data
	if err := r.ParseMultipartForm(h.fileConfig.GetMaxSize("common")); err != nil {
		lib.WriteError(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll()
		}
	}()

	// 2. Получаем spool_id
	spoolIDStr := strings.TrimSpace(r.FormValue("spool_id"))
	if spoolIDStr == "" {
		lib.WriteError(w, "spool_id is required", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	spoolIDUint, err := strconv.ParseUint(spoolIDStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
		return
	}

	// 3. Получаем и нормализуем name
	spoolName := strings.TrimSpace(r.FormValue("name"))
	if spoolName == "" {
		lib.WriteError(w, "spool name is required", http.StatusBadRequest)
		return
	}

	// 4. Проверяем и подготавливаем баннер
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

	// 5. Вызываем usecase
	spool, err := h.usecase.UpdateSpool(r.Context(), usecase.UpdateSpoolInput{
		UserID:      userID,
		SpoolID:     uint(spoolIDUint),
		Name:        spoolName,
		BannerInput: bannerInput,
	})

	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to update spool", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	// 6. Отдаём ответ
	resp := dto.UpdateSpoolResponse{
		SpoolID:    spool.ID,
		Name:       spool.Name,
		BannerLink: spool.BannerLink,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
