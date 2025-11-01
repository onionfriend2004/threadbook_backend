package external

import "context"

// AttemptSendCodeRepoInterface управляет счётчиком попыток отправки кода подтверждения email
type AttemptSendCodeRepoInterface interface {
	// GetSendAttempts возвращает текущее количество попыток отправки кода для пользователя
	GetSendAttempts(ctx context.Context, userID uint) (int, error)

	// IncrementSendAttempts увеличивает счётчик на 1 и устанавливает TTL
	IncrementSendAttempts(ctx context.Context, userID uint) error

	// ResetSendAttempts удаляет счётчик попыток (вызывается после успешной верификации email)
	ResetSendAttempts(ctx context.Context, userID uint) error
}
