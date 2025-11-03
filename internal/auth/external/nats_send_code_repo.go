package external

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
)

type sendCodeRepo struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewSendCodeRepo(nc *nats.Conn) SendCodeRepoInterface {
	// Получаем JetStream context
	js, err := nc.JetStream()
	if err != nil {
		// Логируем ошибку, но продолжаем работу
		// В реальном приложении нужно обработать эту ошибку appropriately
		fmt.Printf("Warning: failed to get JetStream context: %v\n", err)
	}

	return &sendCodeRepo{
		nc: nc,
		js: js,
	}
}

func (r *sendCodeRepo) SendVerifyCodeForUser(eventType int, code int, user *gdomain.User) error {
	if user == nil || user.Email == "" {
		return ErrInvalidUser
	}

	// Определяем subject в зависимости от типа события
	var subject string
	switch eventType {
	case event.UserRegistered:
		subject = "user.registered"
	case event.UserRequestResendVerifyCode:
		subject = "user.code.resend"
	default:
		return fmt.Errorf("%w: %d", ErrUnsupportedEventType, eventType)
	}

	message := gdomain.UserRegisteredEvent{
		Type:     eventType, // Используем переданный eventType
		Code:     code,
		Email:    user.Email,
		Username: user.Username,
		UserID:   int(user.ID),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToSendCode, err)
	}

	// Публикуем в JetStream если доступен, иначе в обычный NATS
	if r.js != nil {
		_, err = r.js.Publish(subject, data)
	} else {
		err = r.nc.Publish(subject, data)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToSendCode, err)
	}

	return nil
}

// Дополнительный метод для явного указания типа события
func (r *sendCodeRepo) SendWelcomeEmail(code int, user *gdomain.User) error {
	return r.SendVerifyCodeForUser(event.UserRegistered, code, user)
}

func (r *sendCodeRepo) SendCodeResend(code int, user *gdomain.User) error {
	return r.SendVerifyCodeForUser(event.UserRequestResendVerifyCode, code, user)
}
