package gdomain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Thread struct {
	ID          uint       `gorm:"primaryKey"`
	CreatorID   uint       `gorm:"not null"`
	SpoolID     *uint      `gorm:"default:NULL"`
	Title       string     `gorm:"not null"`
	Type        ThreadType `gorm:"not null"`
	IsClosed    bool       `gorm:"not null;default:false"`
	AccessLevel uint       `gorm:"not null;default:0"`

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

type ThreadType string

const (
	ThreadTypePublic  ThreadType = "public"
	ThreadTypePrivate ThreadType = "private"
)

// г̶о̶в̶н̶о TODO: Подумать дольше 3х секунд
func (t *ThreadType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "public":
		*t = ThreadTypePublic
	case "private":
		*t = ThreadTypePrivate
	default:
		return fmt.Errorf("invalid thread type: %s", s)
	}
	return nil
}

func (t ThreadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ThreadType) IsValid() bool {
	return t == ThreadTypePublic || t == ThreadTypePrivate
}
