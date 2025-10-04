package external

import "github.com/onionfriend2004/threadbook_backend/internal/gdomain"

type SendCodeRepoInterface interface {
	SendVerifyCodeForUser(code int, user *gdomain.User) error
}
