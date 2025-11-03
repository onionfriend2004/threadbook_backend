package external

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/nats-io/nats.go"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
)

type sendCodeRepo struct {
	nc      *nats.Conn
	subject string
}

func NewSendCodeRepo(nc *nats.Conn, subject string) SendCodeRepoInterface {
	return &sendCodeRepo{
		nc:      nc,
		subject: subject,
	}
}

func (r *sendCodeRepo) SendVerifyCodeForUser(eventType int, code int, user *gdomain.User) error {
	if user == nil || user.Email == "" {
		return ErrInvalidUser
	}

	message := gdomain.EmailEvent{
		Type:  event.UserRegistered,
		Code:  code,
		Email: user.Email,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToSendCode, err)
	}

	if err := r.nc.Publish(r.subject, data); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToSendCode, err)
	}

	return nil
}
