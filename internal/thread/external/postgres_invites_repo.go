package external

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"gorm.io/gorm"
)

type InviteLinkRepoInterface interface {
	CreateLink(ctx context.Context, resourceType string, threadID, creatorID uint, expiresAt time.Time, maxUses uint) (*gdomain.InviteLink, error)
	GetInviteByID(ctx context.Context, inviteID string) (*gdomain.InviteLink, error)
	GetLinksByResource(ctx context.Context, resourceType string, resourceID uint) ([]*gdomain.InviteLink, error)
	DeleteLink(ctx context.Context, inviteID string) error
	DecrementUsage(ctx context.Context, inviteID string) error
	IsValid(ctx context.Context, inviteID string) (bool, error)
}

type inviteLinkRepo struct {
	db *gorm.DB
}

func NewInviteLinkRepo(db *gorm.DB) InviteLinkRepoInterface {
	return &inviteLinkRepo{db: db}
}

func (r *inviteLinkRepo) CreateLink(ctx context.Context, resourceType string, threadID, creatorID uint, expiresAt time.Time, maxUses uint) (*gdomain.InviteLink, error) {
	invite := &gdomain.InviteLink{
		ID:            uuid.New().String(),
		ResourceID:    threadID,
		CreatorID:     creatorID,
		RemainingUses: maxUses,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := r.db.WithContext(ctx).Create(invite).Error
	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (r *inviteLinkRepo) GetInviteByID(ctx context.Context, inviteID string) (*gdomain.InviteLink, error) {
	var invite gdomain.InviteLink
	err := r.db.WithContext(ctx).Where("id = ?", inviteID).First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *inviteLinkRepo) GetLinksByResource(ctx context.Context, resourceType string, resourceID uint) ([]*gdomain.InviteLink, error) {
	var invites []*gdomain.InviteLink
	err := r.db.WithContext(ctx).
		Where("type = ? AND resource_id = ?", resourceType, resourceID).
		Find(&invites).Error
	if err != nil {
		return nil, err
	}
	return invites, nil
}

func (r *inviteLinkRepo) DeleteLink(ctx context.Context, inviteID string) error {
	return r.db.WithContext(ctx).Where("id = ?", inviteID).Delete(&gdomain.InviteLink{}).Error
}

func (r *inviteLinkRepo) DecrementUsage(ctx context.Context, inviteID string) error {
	result := r.db.WithContext(ctx).Model(&gdomain.InviteLink{}).
		Where("id = ? AND remaining_uses > 0", inviteID).
		Updates(map[string]interface{}{
			"remaining_uses": gorm.Expr("remaining_uses - 1"),
			"updated_at":     time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no remaining uses or invite not found")
	}

	return nil
}

func (r *inviteLinkRepo) IsValid(ctx context.Context, inviteID string) (bool, error) {
	var invite gdomain.InviteLink
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires_at > ? AND remaining_uses > 0", inviteID, time.Now()).
		First(&invite).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
