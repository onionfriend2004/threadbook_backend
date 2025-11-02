package gdomain

import "time"

type Message struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	ThreadID  uint      `gorm:"not null;index"` // связь с Thread
	UserID    uint      `gorm:"not null;index"` // автор сообщения
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// связи
	Thread   Thread           `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE"`
	User     User             `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Payloads []MessagePayload `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

type MessagePayload struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	MessageID uint      `gorm:"not null;index"` // связь с Message
	FileLink  string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// связь
	Message Message `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE;"`
}
