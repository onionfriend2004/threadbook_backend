package usecase

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
)
