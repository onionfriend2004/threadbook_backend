package usecase

import (
	"context"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/file/external"
	"go.uber.org/zap"
)

type FileUsecaseInterface interface {
	GetFile(ctx context.Context, input GetFileInput) ([]byte, string, error)
	GetFileWithMetadata(ctx context.Context, input GetFileInput) ([]byte, string, string, time.Time, error)
	SaveFile(ctx context.Context, input SaveFile) (string, error)
	DeleteFile(ctx context.Context, input DeleteFileInput) error
	GetBucketName() string
}

type fileUsecase struct {
	repo   external.FileRepoInterface
	logger *zap.Logger
	Bucket string
}

func NewFileUsecase(repo external.FileRepoInterface, logger *zap.Logger) FileUsecaseInterface {
	return &fileUsecase{
		repo:   repo,
		logger: logger,
		Bucket: repo.GetBucketName(),
	}
}

func (u *fileUsecase) GetBucketName() string {
	return u.Bucket
}

func (u *fileUsecase) GetFile(ctx context.Context, input GetFileInput) ([]byte, string, error) {
	if input.Filename == "" {
		return nil, "", ErrInvalidInput
	}

	data, contentType, err := u.repo.GetFile(ctx, input.Bucket, input.Filename)
	if err != nil {
		u.logger.Error("failed to get file", zap.Error(err))
		return nil, "", ErrFileNotFound
	}

	return data, contentType, nil
}

func (u *fileUsecase) SaveFile(ctx context.Context, input SaveFile) (string, error) {
	fileLink, err := u.repo.SaveFile(ctx, input.Filename, input.File, input.Size, input.ContentType)
	if err != nil {
		return "", err
	}
	return fileLink, nil
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

func (u *fileUsecase) GetFileWithMetadata(ctx context.Context, input GetFileInput) ([]byte, string, string, time.Time, error) {
	if input.Filename == "" {
		return nil, "", "", time.Time{}, ErrInvalidInput
	}

	data, contentType, etag, modTime, err := u.repo.GetFileWithMetadata(ctx, input.Bucket, input.Filename)
	if err != nil {
		u.logger.Error("failed to get file with metadata", zap.Error(err))
		return nil, "", "", time.Time{}, ErrFileNotFound
	}

	return data, contentType, etag, modTime, nil
}
