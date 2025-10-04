package external

import "errors"

var (
	ErrInvalidEmail           = errors.New("invalid email address")
	ErrNoMXRecords            = errors.New("no MX records found")
	ErrSMTPConnect            = errors.New("failed to connect to SMTP server")
	ErrEmailNotExists         = errors.New("email address does not exist")
	ErrSMTPCommand            = errors.New("SMTP command failed")
	ErrDialMX                 = errors.New("failed to dial MX server")
	ErrSMTPClientCreate       = errors.New("failed to create SMTP client for verification")
	ErrVerificationMailFailed = errors.New("MAIL FROM command failed during verification")
	ErrVerificationRcptFailed = errors.New("RCPT TO command failed during verification")
)
