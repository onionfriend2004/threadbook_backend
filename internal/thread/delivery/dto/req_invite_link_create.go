package dto

import "time"

type CreateInviteLinkRequest struct {
	MaxUses   uint      `json:"max_uses" validate:"required,gte=1"`
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
}
