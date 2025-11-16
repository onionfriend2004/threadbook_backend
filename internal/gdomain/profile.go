package gdomain

import (
	"time"
)

type Profile struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	UserID     uint      `gorm:"uniqueIndex;not null"`
	Nickname   string    `gorm:"size:32"`
	AvatarLink string    `gorm:"size:255"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
