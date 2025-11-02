package lib

import (
	"errors"
)

var (
	ErrInvalidRequestData = errors.New("invalid request")
	ErrInternalServer     = errors.New("internal server error")
)
