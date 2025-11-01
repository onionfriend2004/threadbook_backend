package gdomain

import (
	"strings"
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex:idx_users_email;not null"`
	EmailVerify  bool      `gorm:"not null"`
	Username     string    `gorm:"uniqueIndex:idx_users_username;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func NewUser(email, username, passwordHash string) (*User, error) {
	if email == "" || username == "" || passwordHash == "" {
		return nil, ErrInvalidUser
	}

	return &User{
		Email:        NormalizeEmail(email),
		Username:     NormalizeUsername(username),
		PasswordHash: passwordHash,
	}, nil
}
