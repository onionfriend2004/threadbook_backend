package dto

import "time"

type ThreadCreateResponse struct {
	ID        uint      `json:"id"`
	SpoolID   uint      `json:"spool_id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	IsClosed  bool      `json:"is_closed"`
	IsCreator bool      `json:"is_creator"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
