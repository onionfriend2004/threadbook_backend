package gdomain

import (
	"time"
)

type Thread struct {
	ID          uint   `gorm:"primaryKey"`
	CreatorID   uint   `gorm:"not null"`
	SpoolID     *uint  `gorm:"default:NULL"`
	Title       string `gorm:"not null"`
	Type        string `gorm:"not null"`
	IsClosed    bool   `gorm:"not null;default:false"`
	AccessLevel uint   `gorm:"nut null;default:1"`

	Messages []Message `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE;"`
	Users    []User    `gorm:"many2many:thread_users;constraint:OnDelete:CASCADE;"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type ThreadUser struct {
	UserID   uint             `gorm:"primaryKey"`
	ThreadID uint             `gorm:"primaryKey"`
	Status   ThreadUserStatus `gorm:"type:varchar(20);not null;default:'active'"`
	JoinedAt time.Time        `gorm:"autoCreateTime"`
}

type ThreadUserStatus string

const (
	ThreadUserStatusActive ThreadUserStatus = "active"
	ThreadUserStatusLeft   ThreadUserStatus = "left"
	ThreadUserStatusBanned ThreadUserStatus = "banned"
	ThreadUserStatusMuted  ThreadUserStatus = "muted"
)
