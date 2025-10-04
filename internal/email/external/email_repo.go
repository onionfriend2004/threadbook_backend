package external

import "net/smtp"

type MailRepositoryInterface interface {
	Send(to, msg string) error
	VerifyEmail(email string) (bool, error)

	OpenConnection() (*smtp.Client, error)
	CloseConnection() error
}
