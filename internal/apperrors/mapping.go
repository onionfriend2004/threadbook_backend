package apperrors

import (
	"errors"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"

	// Spool
	spoolExternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	spoolUsecase "github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"

	// Thread
	threadExternal "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	threadUsecase "github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"

	// Auth
	authExternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	authUsecase "github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"

	// Domain
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

var errToCode = map[error]int{
	// --- Общие ошибки запроса ---
	lib.ErrInvalidRequestData:           http.StatusBadRequest,
	gdomain.ErrInvalidUser:              http.StatusBadRequest,
	gdomain.ErrEmptyName:                http.StatusBadRequest,
	spoolUsecase.ErrInvalidInput:        http.StatusBadRequest,
	threadUsecase.ErrInvalidInput:       http.StatusBadRequest,
	spoolExternal.ErrInvalidSpool:       http.StatusBadRequest,
	threadExternal.ErrInvalidThreadType: http.StatusBadRequest,
	threadUsecase.ErrWrongTypeThread:    http.StatusBadRequest,
	authUsecase.ErrInvalidInput:         http.StatusBadRequest,
	authExternal.ErrInvalidSessionData:  http.StatusBadRequest,

	// --- Ошибки 401 Unauthorized ---
	gdomain.ErrUnauthorized:           http.StatusUnauthorized,
	authUsecase.ErrInvalidCredentials: http.StatusUnauthorized,

	// --- Ошибки 403 Forbidden (нет доступа / слишком много попыток) ---
	spoolUsecase.ErrForbidden:           http.StatusForbidden,
	threadExternal.ErrPermissionDenied:  http.StatusForbidden,
	threadExternal.ErrUserNoAccess:      http.StatusForbidden,
	threadUsecase.ErrNoAccessToThread:   http.StatusForbidden,
	threadUsecase.ErrNoRightsOnJoinRoom: http.StatusForbidden,
	threadUsecase.ErrThreadIsClosed:     http.StatusForbidden,
	threadUsecase.ErrFailedToEnsureRoom: http.StatusForbidden,
	authUsecase.ErrTooManyAttempts:      http.StatusForbidden,

	// --- Ошибки 404 Not Found ---
	spoolExternal.ErrNotFound:          http.StatusNotFound,
	spoolExternal.ErrUserNotFound:      http.StatusNotFound,
	threadExternal.ErrThreadNotFound:   http.StatusNotFound,
	threadExternal.ErrUserNotInSpool:   http.StatusNotFound,
	threadExternal.ErrMessageNotFound:  http.StatusNotFound,
	threadUsecase.ErrThreadNotFound:    http.StatusNotFound,
	threadUsecase.ErrFailedToGetThread: http.StatusNotFound,
	authUsecase.ErrUserNotFound:        http.StatusNotFound,
	authUsecase.ErrSessionNotFound:     http.StatusNotFound,
	authExternal.ErrSessionNotFound:    http.StatusNotFound,
	gdomain.ErrNotFound:                http.StatusNotFound,

	// --- Ошибки 409 Conflict ---
	spoolExternal.ErrSpoolExists:        http.StatusConflict,
	spoolExternal.ErrUserAlreadyInSpool: http.StatusConflict,
	gdomain.ErrUserExists:               http.StatusConflict,
	authUsecase.ErrUserAlreadyExists:    http.StatusConflict,
	authExternal.ErrUserExists:          http.StatusConflict,
	authUsecase.ErrAlreadyConfirmed:     http.StatusConflict,

	// --- Ошибки 500 Internal Server Error ---
	spoolUsecase.ErrInternal:            http.StatusInternalServerError,
	spoolUsecase.ErrFailedToSaveBanner:  http.StatusInternalServerError,
	spoolUsecase.ErrFailedToCreateSpool: http.StatusInternalServerError,
	spoolUsecase.ErrFailedToUpdateSpool: http.StatusInternalServerError,
	spoolUsecase.ErrFailedToInvite:      http.StatusInternalServerError,
	spoolUsecase.ErrFailedToGetMembers:  http.StatusInternalServerError,
	spoolUsecase.ErrFailedToGetSpool:    http.StatusInternalServerError,

	threadExternal.ErrMarshalPublishData:     http.StatusInternalServerError,
	threadExternal.ErrCentrifugoPublish:      http.StatusInternalServerError,
	threadExternal.ErrGenerateConnectToken:   http.StatusInternalServerError,
	threadExternal.ErrGenerateSubscribeToken: http.StatusInternalServerError,
	threadExternal.ErrMessageNil:             http.StatusInternalServerError,
	threadExternal.ErrCreateMessage:          http.StatusInternalServerError,
	threadExternal.ErrCreatePayloads:         http.StatusInternalServerError,
	threadExternal.ErrGetMessages:            http.StatusInternalServerError,
	threadExternal.ErrGetMessageByID:         http.StatusInternalServerError,
	threadExternal.ErrDeleteMessage:          http.StatusInternalServerError,
	threadExternal.ErrCountMessages:          http.StatusInternalServerError,
	threadExternal.ErrInviteFailed:           http.StatusInternalServerError,
	threadExternal.ErrCreateThread:           http.StatusInternalServerError,
	threadExternal.ErrGetThreads:             http.StatusInternalServerError,
	threadExternal.ErrCloseThread:            http.StatusInternalServerError,
	threadExternal.ErrUpdateThread:           http.StatusInternalServerError,
	threadExternal.ErrGetMembers:             http.StatusInternalServerError,
	threadExternal.ErrGetAccessibleIDs:       http.StatusInternalServerError,
	threadExternal.ErrRightsCheck:            http.StatusInternalServerError,

	threadUsecase.ErrFailToCreateVoiceToken: http.StatusInternalServerError,
	threadUsecase.ErrFailedToSaveMsg:        http.StatusInternalServerError,
	threadUsecase.ErrFailedToPublish:        http.StatusInternalServerError,

	authExternal.ErrFailedToSendCode: http.StatusInternalServerError,

	lib.ErrInternalServer: http.StatusInternalServerError,
}

func GetErrAndCodeToSend(err error) (int, error) {
	for knownErr, code := range errToCode {
		if errors.Is(err, knownErr) {
			return code, knownErr
		}
	}
	return http.StatusInternalServerError, lib.ErrInternalServer
}
