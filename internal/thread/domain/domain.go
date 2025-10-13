package domain

import "time"

type Thread struct {
	ID        int       `gorm:"column:id;primaryKey"`
	CreatorID int       `gorm:"column:creator_id;not null"`
	SpoolID   int       `gorm:"column:spool_id;not null"`
	Title     string    `gorm:"column:title;not null"`
	Type      string    `gorm:"column:type;not null"`
	IsClosed  bool      `gorm:"column:is_closed;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

type UpdateThreadInput struct {
	ID       int     `json:"id"`
	EditorID int     `json:"editor_id"`
	Title    *string `json:"title"`
	Type     *string `json:"type"`
	IsClosed *bool   `json:"is_closed"`
}
