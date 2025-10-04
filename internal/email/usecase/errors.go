package usecase

import "errors"

var (
	ErrEmptyEmail           = errors.New("recipient email is empty")
	ErrUnsupportedEmailType = errors.New("unsupported email operation type")
	ErrFailedToSendEmail    = errors.New("failed to send email")
)
