package external

import "errors"

var (
	ErrThreadNotFound   = errors.New("thread not found")
	ErrUserNotInSpool   = errors.New("user not in spool")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUserNoAccess     = errors.New("user not owner")
)
