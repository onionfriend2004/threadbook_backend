package deliveryHTTP

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"go.uber.org/zap"
)

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	filename, err := url.PathUnescape(filename)
	if err != nil {
		lib.WriteError(w, "invalid file path", http.StatusBadRequest)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	bucket, err = url.PathUnescape(bucket)
	if err != nil {
		lib.WriteError(w, "invalid file path", http.StatusBadRequest)
		return
	}

	input := usecase.GetFileInput{Filename: filename, Bucket: bucket}
	if input.Filename == "" || input.Bucket == "" {
		lib.WriteError(w, "filename required", http.StatusBadRequest)
		return
	}

	data, contentType, err := h.usecase.GetFile(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to get file", zap.Error(err))
		switch err {
		case usecase.ErrInvalidInput:
			lib.WriteError(w, err.Error(), http.StatusBadRequest)
		case usecase.ErrFileNotFound:
			lib.WriteError(w, err.Error(), http.StatusNotFound)
		default:
			lib.WriteError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		h.logger.Warn("failed to write response", zap.Error(err))
	}
}
