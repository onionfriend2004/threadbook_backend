package gdomain

import (
	"strings"
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	Email        string    `gorm:"uniqueIndex:idx_users_email" json:"email"`
	EmailVerify  bool      `gorm:"not null" json:"is_verify"`
	Username     string    `gorm:"uniqueIndex:idx_users_username;not null" json:"username"`
	Profile      Profile   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID"`
	PasswordHash string    `gorm:"" json:"-"`
	IsGuest      bool      `gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func NewUser(email, username, passwordHash string) (*User, error) {
	if username == "" || passwordHash == "" {
		return nil, ErrInvalidUser
	}

	return &User{
		Email:        NormalizeEmail(email),
		Username:     NormalizeUsername(username),
		PasswordHash: passwordHash,
	}, nil
}
