package domain

import "time"

type NoAuthSession struct {
	ID        string    `json:"id"`
	ThreadID  uint      `json:"thread_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewSession(id string, threadID uint, expiresAt time.Time) *NoAuthSession {
	return &NoAuthSession{
		ID:        id,
		ThreadID:  threadID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
}
