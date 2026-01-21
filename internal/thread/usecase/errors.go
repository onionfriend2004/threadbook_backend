package usecase

import "errors"

var (
	// Общие ошибки
	ErrInvalidInput   = errors.New("invalid input")
	ErrThreadNotFound = errors.New("thread not found")

	// Ошибки, связанные с комнатами / правами
	ErrFailedToEnsureRoom     = errors.New("failed to ensure room")
	ErrNoRightsOnJoinRoom     = errors.New("no rights to join thread room")
	ErrWrongTypeThread        = errors.New("wrong type of thread")
	ErrFailToCreateVoiceToken = errors.New("failed to create voice token")
	ErrFailToGetThread        = errors.New("failed to get thread in room usecase")

	// Ошибки сообщений
	ErrNoAccessToThread         = errors.New("user has no access to this thread")
	ErrNoAccessToMessage        = errors.New("user has no access to this message")
	ErrThreadIsClosed           = errors.New("cannot send message: thread is closed")
	ErrFailedToSaveMsg          = errors.New("failed to save message")
	ErrFailedToGetThread        = errors.New("failed to get thread")
	ErrFailedToPublish          = errors.New("failed to publish message event")
	ErrMessageNotFound          = errors.New("message not found")
	ErrFailedToUpdateMsg        = errors.New("failed to update message")
	ErrFailedToDeleteMsg        = errors.New("failed to delete message")
	ErrFailedToSaveMessageFiles = errors.New("failed to save message files")
	ErrFailedToRollbackFiles    = errors.New("failed to rollback uploaded files")
	ErrMessageCannotBeEmpty     = errors.New("message cannot be empty")

	ErrThreadClosed   = errors.New("thread is closed")
	ErrSessionExpired = errors.New("session expired")
)
