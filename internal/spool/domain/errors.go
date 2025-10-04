package domain

import "errors"

var (
	ErrEmptyName = errors.New("spool name cannot be empty")
)
