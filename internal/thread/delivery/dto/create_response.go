package dto

import "time"

type ThreadCreateResponse struct {
	ID        int       `json:"id"`
	SpoolID   int       `json:"spool_id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	IsClosed  bool      `json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
