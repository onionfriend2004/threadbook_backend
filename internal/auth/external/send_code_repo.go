package external

import "github.com/onionfriend2004/threadbook_backend/internal/gdomain"

type SendCodeRepoInterface interface {
	SendVerifyCodeForUser(eventType int, code int, user *gdomain.User) error
	SendWelcomeEmail(code int, user *gdomain.User) error
	SendCodeResend(code int, user *gdomain.User) error
}
