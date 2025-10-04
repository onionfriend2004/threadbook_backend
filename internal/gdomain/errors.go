package gdomain

import "errors"

var (
	ErrInvalidUser  = errors.New("invalid user data")
	ErrUserExists   = errors.New("user already exists")
	ErrNotFound     = errors.New("user not found")
	ErrUnauthorized = errors.New("invalid credentials")
)
