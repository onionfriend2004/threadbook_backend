package gdomain

import (
	"time"
)

type Thread struct {
	ID        int       `gorm:"primaryKey;column:id"`
	SpoolID   int       `gorm:"column:spool_id;not null"`
	Title     string    `gorm:"column:title;not null"`
	Type      string    `gorm:"column:type;not null"`
	IsClosed  bool      `gorm:"column:is_closed;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Thread) TableName() string {
	return "threads"
}
