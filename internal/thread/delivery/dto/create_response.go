package dto

import (
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type ThreadCreateResponse struct {
	ID          uint               `json:"id"`
	SpoolID     uint               `json:"spool_id"`
	AccessLevel uint               `json:"access_level"`
	Title       string             `json:"title"`
	Type        gdomain.ThreadType `json:"type"`
	IsClosed    bool               `json:"is_closed"`
	IsCreator   bool               `json:"is_creator"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}
