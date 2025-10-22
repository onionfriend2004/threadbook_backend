package dto

import "time"

type MessageResponse struct {
	ID        uint      `json:"id"`
	ThreadID  uint      `json:"thread_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Payloads  []any     `json:"payloads,omitempty"` // можно заменить на конкретный тип
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
