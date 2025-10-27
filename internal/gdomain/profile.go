package gdomain

import (
	"time"
)

type Profile struct {
	ID         int    `gorm:"primaryKey;autoIncrement"`
	UserID     int    `gorm:"uniqueIndex;not null"`
	Nickname   string `gorm:"size:32"`
	AvatarLink string `gorm:"size:255"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
