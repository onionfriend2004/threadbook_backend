package gdomain

import (
	"time"
)

type Thread struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	CreatorID uint      `gorm:"column:creator_id;not null"`
	SpoolID   uint      `gorm:"column:spool_id;not null"`
	Title     string    `gorm:"column:title;not null"`
	Type      string    `gorm:"column:type;not null"`
	IsClosed  bool      `gorm:"column:is_closed;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Messages []Message `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE;"`

	Users []User `gorm:"many2many:thread_users;joinForeignKey:ThreadID;joinReferences:UserID;constraint:OnDelete:CASCADE;"`
}

type ThreadUser struct {
	UserID   uint `gorm:"primaryKey"`
	ThreadID uint `gorm:"primaryKey"`
	IsMember bool `gorm:"default:true"`
}
