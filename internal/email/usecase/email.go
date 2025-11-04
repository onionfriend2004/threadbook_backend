package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/onionfriend2004/threadbook_backend/internal/email/external"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"go.uber.org/zap"
)

type EmailUsecaseInterface interface {
	SendWelcomeEmail(userRegisteredEvent *gdomain.UserRegisteredEvent) error
	SendCodeResendEmail(userRegisteredEvent *gdomain.UserRegisteredEvent) error
	// SendMessageOnEmail(userRegisteredEvent *gdomain.UserRegisteredEvent) error // !!!deprecated!!!
}

func isPermanentError(err error) bool {
	return errors.Is(err, external.ErrInvalidEmail) ||
		errors.Is(err, external.ErrNoMXRecords) ||
		errors.Is(err, external.ErrVerificationRcptFailed)
}

type emailUseCase struct {
	emailRepo external.MailRepositoryInterface
	logger    *zap.Logger
}

func NewEmailUsecase(emailRepo external.MailRepositoryInterface, logger *zap.Logger) EmailUsecaseInterface {
	return &emailUseCase{emailRepo: emailRepo, logger: logger}
}
func (e *emailUseCase) SendWelcomeEmail(userRegisteredEvent *gdomain.UserRegisteredEvent) error {
	if userRegisteredEvent.Email == "" {
		return ErrPermanentEmailError
	}

	if userRegisteredEvent.Type != event.UserRegistered {
		return ErrPermanentEmailError
	}

	safeEmail := sanitizeHeader(userRegisteredEvent.Email)
	subject := "Welcome to ThreadBook - Verify Your Email"
	body := fmt.Sprintf(`
		<h1>Welcome to ThreadBook, %s!</h1>
		<p>Your verification code is: <strong>%d</strong></p>
		<p>Use this code to complete your registration.</p>
	`, userRegisteredEvent.Username, userRegisteredEvent.Code)

	safeSubject := sanitizeHeader(subject)
	msg := formatMessage(safeEmail, safeSubject, body)

	if err := e.emailRepo.Send(userRegisteredEvent.Email, msg); err != nil {
		if isPermanentError(err) {
			e.logger.Warn("permanent email failure, will not retry",
				zap.String("email", userRegisteredEvent.Email),
				zap.Error(err))
			return ErrPermanentEmailError
		}

		e.logger.Error("temporary email failure, will retry",
			zap.String("email", userRegisteredEvent.Email),
			zap.Error(err))
		return fmt.Errorf("%w: %v", ErrFailedToSendEmail, err)
	}

	e.logger.Info("welcome email sent successfully",
		zap.String("email", userRegisteredEvent.Email))
	return nil
}

func (e *emailUseCase) SendCodeResendEmail(userRegisteredEvent *gdomain.UserRegisteredEvent) error {
	if userRegisteredEvent.Email == "" {
		return ErrPermanentEmailError
	}

	if userRegisteredEvent.Type != event.UserRequestResendVerifyCode {
		return ErrPermanentEmailError
	}

	safeEmail := sanitizeHeader(userRegisteredEvent.Email)
	subject := "ThreadBook - Your Verification Code"
	body := fmt.Sprintf(`
		<h1>Your Verification Code</h1>
		<p>Your verification code is: <strong>%d</strong></p>
		<p>If you didn't request this code, please ignore this email.</p>
	`, userRegisteredEvent.Code)

	safeSubject := sanitizeHeader(subject)
	msg := formatMessage(safeEmail, safeSubject, body)

	if err := e.emailRepo.Send(userRegisteredEvent.Email, msg); err != nil {
		if isPermanentError(err) {
			e.logger.Warn("permanent email failure in resend, will not retry",
				zap.String("email", userRegisteredEvent.Email),
				zap.Error(err))
			return ErrPermanentEmailError
		}

		e.logger.Error("temporary email failure in resend, will retry",
			zap.String("email", userRegisteredEvent.Email),
			zap.Error(err))
		return fmt.Errorf("%w: %v", ErrFailedToSendEmail, err)
	}

	e.logger.Info("code resend email sent successfully",
		zap.String("email", userRegisteredEvent.Email))
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
