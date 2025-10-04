package external

import (
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// WARNING !!!
// LEGACY CODE =))) HAHAHH
// WARNING !!!

type MailRepository struct {
	smtpClient *smtp.Client

	from_sender string
	smtpServer  string
	smtpPort    string
	username    string
	password    string
}

func NewMailRepository(smtpServer, smtpPort, username, password, from_sender string) MailRepositoryInterface {
	mailRepo := &MailRepository{
		from_sender: from_sender,
		smtpServer:  smtpServer,
		smtpPort:    smtpPort,
		username:    username,
		password:    password,
	}

	if err := mailRepo.ensureConnected(); err != nil {
		log.Fatalf("Failed to connect to SMTP server: %v", err)
	}

	return mailRepo
}

func (r *MailRepository) Send(to, msg string) error {
	if err := r.ensureConnected(); err != nil {
		return err
	}

	isExist, err := r.VerifyEmail(to)
	if !isExist || err != nil {
		return err
	}
	if err := r.smtpClient.Mail(r.from_sender); err != nil {
		return err
	}
	if err := r.smtpClient.Rcpt(to); err != nil {
		return err
	}
	w, err := r.smtpClient.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

func (r *MailRepository) VerifyEmail(email string) (bool, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, ErrInvalidEmail
	}
	domain := parts[1]

	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return false, ErrNoMXRecords
	}

	host := mxRecords[0].Host

	conn, err := net.DialTimeout("tcp", host+":25", 10*time.Second)
	if err != nil {
		return false, ErrDialMX
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return false, ErrSMTPClientCreate
	}
	defer func() {
		if err := client.Quit(); err != nil {
			log.Printf("Error quitting verification client: %v", err)
		}
	}()

	if err := client.Mail(r.from_sender); err != nil {
		return false, ErrVerificationMailFailed
	}
	if err := client.Rcpt(email); err != nil {
		return false, ErrVerificationRcptFailed
	}

	return true, nil
}

func (r *MailRepository) OpenConnection() (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         r.smtpServer,
	}

	conn, err := tls.Dial("tcp", r.smtpServer+":"+r.smtpPort, tlsConfig)
	if err != nil {
		return nil, err
	}

	smtpClient, err := smtp.NewClient(conn, r.smtpServer)
	if err != nil {
		return nil, err
	}

	err = smtpClient.Auth(smtp.PlainAuth("", r.username, r.password, r.smtpServer))
	if err != nil {
		return nil, err
	}

	return smtpClient, nil
}

func (r *MailRepository) CloseConnection() error {
	return r.smtpClient.Quit()
}

func (r *MailRepository) ensureConnected() error {
	if r.smtpClient != nil {
		if err := r.smtpClient.Noop(); err == nil {
			return nil
		}

		if err := r.smtpClient.Quit(); err != nil {
			log.Printf("Error closing old SMTP connection: %v", err)
		}
		r.smtpClient = nil
	}

	client, err := r.OpenConnection()
	if err != nil {
		return err
	}
	r.smtpClient = client
	return nil
}

var _ MailRepositoryInterface = (*MailRepository)(nil)
