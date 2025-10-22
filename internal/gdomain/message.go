package gdomain

import "time"

type Message struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ThreadID  int       `gorm:"not null;index" json:"thread_id"` // связь с Thread
	UserID    int       `gorm:"not null;index" json:"user_id"`   // автор сообщения
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// связи
	Thread   Thread           `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE" json:"thread,omitempty"`
	User     User             `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Payloads []MessagePayload `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"payloads,omitempty"`
}

type MessagePayload struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"` // связь с Message
	FileLink  string    `gorm:"type:text;not null" json:"file_link"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// связь
	Message Message `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE;" json:"message,omitempty"`
}
