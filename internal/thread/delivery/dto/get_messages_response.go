package dto

import "time"

type MessageResponse struct {
	ID        uint      `json:"id"`
	ThreadID  uint      `json:"thread_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Payloads  []any     `json:"payloads,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
