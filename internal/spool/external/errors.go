package external

import "errors"

var (
	ErrInvalidSpool = errors.New("invalid spool")
	ErrSpoolExists  = errors.New("spool already exists")
	ErrNotFound     = errors.New("record not found")
)
