package gdomain

import (
	"time"
)

type Thread struct {
	ID        int       `gorm:"column:id;primaryKey"`
	CreatorID int       `gorm:"column:creator_id;not null"`
	SpoolID   int       `gorm:"column:spool_id;not null"`
	Title     string    `gorm:"column:title;not null"`
	Type      string    `gorm:"column:type;not null"`
	IsClosed  bool      `gorm:"column:is_closed;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Spool Spool  `gorm:"foreignKey:SpoolID;references:ID"`
	Users []User `gorm:"many2many:thread_users;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Thread) TableName() string {
	return "threads"
}
