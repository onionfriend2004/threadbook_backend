package usecase

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/file/external"
	"go.uber.org/zap"
)

type FileUsecaseInterface interface {
	GetFile(ctx context.Context, input GetFileInput) ([]byte, string, error)
	SaveFile(ctx context.Context, input SaveFileInput) error
	DeleteFile(ctx context.Context, input DeleteFileInput) error
}

type fileUsecase struct {
	repo   external.FileRepoInterface
	logger *zap.Logger
}

func NewFileUsecase(repo external.FileRepoInterface, logger *zap.Logger) FileUsecaseInterface {
	return &fileUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *fileUsecase) GetFile(ctx context.Context, input GetFileInput) ([]byte, string, error) {
	if input.Filename == "" {
		return nil, "", ErrInvalidInput
	}

	data, contentType, err := u.repo.GetFile(ctx, input.Filename)
	if err != nil {
		u.logger.Error("failed to get file", zap.Error(err))
		return nil, "", ErrFileNotFound
	}

	return data, contentType, nil
}

func (u *fileUsecase) SaveFile(ctx context.Context, input SaveFileInput) error {
	if input.Filename == "" || len(input.Data) == 0 {
		return ErrInvalidInput
	}

	if err := u.repo.SaveFile(ctx, input.Filename, input.Data, input.ContentType); err != nil {
		u.logger.Error("failed to save file", zap.Error(err))
		return ErrSaveFailed
	}

	return nil
}

func (u *fileUsecase) DeleteFile(ctx context.Context, input DeleteFileInput) error {
	if input.Filename == "" {
		return ErrInvalidInput
	}

	if err := u.repo.DeleteFile(ctx, input.Filename); err != nil {
		u.logger.Error("failed to delete file", zap.Error(err))
		return ErrDeleteFailed
	}

	return nil
}
