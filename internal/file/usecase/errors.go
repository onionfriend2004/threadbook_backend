package usecase

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrFileNotFound = errors.New("file not found")
	ErrSaveFailed   = errors.New("failed to save file")
	ErrDeleteFailed = errors.New("failed to delete file")
)
