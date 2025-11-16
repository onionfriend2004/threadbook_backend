package gdomain

import "time"

type InviteLink struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Type          string    `gorm:"not null" json:"type"`        // "thread" или "spool"
	ResourceID    uint      `gorm:"not null" json:"resource_id"` // thread_id или spool_id
	CreatorID     uint      `gorm:"not null" json:"creator_id"`
	RemainingUses uint      `gorm:"not null;default:1" json:"remaining_uses"`
	ExpiresAt     time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// func NewSession(id string, threadID uint, expiresAt time.Time) *InviteLink {
// 	return &InviteLink{
// 		ID:        id,
// 		ThreadID:  threadID,
// 		CreatedAt: time.Now(),
// 		ExpiresAt: expiresAt,
// 	}
// }
