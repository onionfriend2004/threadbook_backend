package deliveryHTTP

import (
	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"go.uber.org/zap"
)

type FileHandler struct {
	usecase usecase.FileUsecaseInterface
	logger  *zap.Logger
}

func NewFileHandler(u usecase.FileUsecaseInterface, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		usecase: u,
		logger:  logger,
	}
}

func (h *FileHandler) Routes(r chi.Router) {
	r.Get("/files/{filename}", h.GetFile)
}
