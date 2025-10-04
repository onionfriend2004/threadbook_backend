package usecase

import (
	"fmt"
	"strings"

	"github.com/onionfriend2004/threadbook_backend/internal/email/external"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"go.uber.org/zap"
)

type EmailUsecaseInterface interface {
	SendMessageOnEmail(emailEvent *gdomain.EmailEvent) error
}

type emailUseCase struct {
	emailRepo external.MailRepositoryInterface
	logger    *zap.Logger
}

func NewEmailUsecase(emailRepo external.MailRepositoryInterface, logger *zap.Logger) EmailUsecaseInterface {
	return &emailUseCase{emailRepo: emailRepo, logger: logger}
}

func (e *emailUseCase) SendMessageOnEmail(emailEvent *gdomain.EmailEvent) error {
	if emailEvent.Email == "" {
		return ErrEmptyEmail
	}

	safeEmail := sanitizeHeader(emailEvent.Email)

	var subject, body string
	switch emailEvent.Type {
	case event.SendVerificationCode:
		subject = "Verification Code"
		body = fmt.Sprintf("<p>Your verification code is: <strong>%d</strong></p>", emailEvent.Code)
	default:
		return fmt.Errorf("%w: %d", ErrUnsupportedEmailType, emailEvent.Type)
	}

	safeSubject := sanitizeHeader(subject)
	msg := formatMessage(safeEmail, safeSubject, body)

	if err := e.emailRepo.Send(emailEvent.Email, msg); err != nil {
		e.logger.Error("failed to send email",
			zap.Int("operation", emailEvent.Type),
			zap.String("email", emailEvent.Email),
			zap.Error(err))
		return fmt.Errorf("%w: %v", ErrFailedToSendEmail, err)
	}

	e.logger.Info("email sent successfully",
		zap.Int("operation", emailEvent.Type),
		zap.String("email", emailEvent.Email))

	return nil
}

func formatMessage(to, subject, body string) string {
	return fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n%s", to, subject, body)
}

func sanitizeHeader(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}

var _ EmailUsecaseInterface = (*emailUseCase)(nil)
