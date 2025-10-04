package external

import "errors"

var (
	ErrInvalidSpool       = errors.New("invalid spool data")
	ErrSpoolExists        = errors.New("spool already exists")
	ErrNotFound           = errors.New("record not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyInSpool = errors.New("user already in spool")
)
