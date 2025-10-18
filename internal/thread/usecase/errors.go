package usecase

import "errors"

var (
	ErrThreadNotFound = errors.New("thread not found")
	ErrInvalidInput   = errors.New("invalid input")

	ErrFaildToEnsureRoom  = errors.New("faild to ensure room")
	ErrNoRightsOnJoinRoom = errors.New("no rights to join thread room")
)
