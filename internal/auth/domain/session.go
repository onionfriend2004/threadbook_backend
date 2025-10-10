package domain

import "time"

type Session struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewSession(id string, userID uint, username string, expiresAt time.Time) *Session {
	return &Session{
		ID:        id,
		UserID:    userID,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
}
