package apperrors

import (
	"errors"
	"net/http"

	authUsecase "github.com/onionfriend2004/threadbook_backend/internal/auth/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	spoolUsecase "github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	threadUsecase "github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
)

var errToCode = map[error]int{
	// HTTP/server level
	lib.ErrInvalidRequestData: http.StatusBadRequest, // 400

	// --- Ошибки spool ---
	spoolUsecase.ErrInvalidInput: http.StatusBadRequest, // 400 — некорректный ввод
	spoolUsecase.ErrNotFound:     http.StatusNotFound,   // 404 — не найден
	spoolUsecase.ErrForbidden:    http.StatusForbidden,  // 403 — доступ запрещён

	// --- Ошибки thread ---
	threadUsecase.ErrThreadNotFound:     http.StatusNotFound,            // 404 — поток не найден
	threadUsecase.ErrInvalidInput:       http.StatusBadRequest,          // 400 — некорректные входные данные
	threadUsecase.ErrFaildToEnsureRoom:  http.StatusInternalServerError, // 500 — ошибка при создании/проверке комнаты
	threadUsecase.ErrNoRightsOnJoinRoom: http.StatusForbidden,           // 403 — нет прав для входа в комнату потока
	threadUsecase.ErrWrognTypeThread:    http.StatusBadRequest,          // 400 — неверный тип потока

	// --- Ошибки auth ---
	authUsecase.ErrUserNotFound:       http.StatusNotFound,     // 404 — пользователь не найден
	authUsecase.ErrSessionNotFound:    http.StatusNotFound,     // 404 — сессия не найдена
	authUsecase.ErrUserAlreadyExists:  http.StatusConflict,     // 409 — пользователь уже существует
	authUsecase.ErrInvalidCredentials: http.StatusUnauthorized, // 401 — неверный логин или пароль
	authUsecase.ErrInvalidInput:       http.StatusBadRequest,   // 400 — некорректные входные данные
	authUsecase.ErrCodeIncorrect:      http.StatusBadRequest,   // 400 — неверный код подтверждения
	authUsecase.ErrTooManyAttempts:    http.StatusForbidden,    // 403 — слишком много попыток отправки кода подтверждения
	authUsecase.ErrAlreadyConfirmed:   http.StatusBadRequest,   // 400 — почта уже подтверждена
}

func GetErrAndCodeToSend(err error) (int, error) {
	for knownErr, code := range errToCode {
		if errors.Is(err, knownErr) {
			return code, knownErr
		}
	}
	return http.StatusInternalServerError, lib.ErrInternalServer
}
