package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external/domain"
	"github.com/redis/go-redis/v9"
)

type InviteLinkRepoInterface interface {
	CreateLink(ctx context.Context, threadID uint) (*domain.InviteLink, error)
	GetInviteByID(ctx context.Context, inviteID string) (*domain.InviteLink, error)
	GetLinksByThread(ctx context.Context, threadID uint) ([]*domain.InviteLink, error)
	DeleteLink(ctx context.Context, inviteID string) error
}

type inviteLink struct {
	redisClient *redis.Client
}

func NewInviteLinkRepo(redisClient *redis.Client) InviteLinkRepoInterface {
	return &inviteLink{
		redisClient: redisClient,
	}
}

func (r *inviteLink) CreateLink(ctx context.Context, threadID uint) (*domain.InviteLink, error) {
	inviteID := uuid.NewString()
	key := "invite_link:" + inviteID // ← меняем ключ
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	invite := &domain.InviteLink{
		ID:        inviteID,
		ThreadID:  threadID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(invite)
	if err != nil {
		return nil, err
	}

	pipe := r.redisClient.TxPipeline()

	// Основная запись: invite_id -> invite_data
	pipe.Set(ctx, key, data, 24*time.Hour)

	// Обратный индекс: thread_id -> [invite_ids]
	threadKey := fmt.Sprintf("thread_invites:%d", threadID) // ← меняем ключ
	pipe.SAdd(ctx, threadKey, inviteID)
	pipe.Expire(ctx, threadKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return invite, nil
}

// Получить invite по ID
func (r *inviteLink) GetInviteByID(ctx context.Context, inviteID string) (*domain.InviteLink, error) {
	key := "invite_link:" + inviteID // ← меняем ключ
	data, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var invite domain.InviteLink
	err = json.Unmarshal(data, &invite)
	return &invite, err
}

// Получить все invites для треда
func (r *inviteLink) GetLinksByThread(ctx context.Context, threadID uint) ([]*domain.InviteLink, error) {
	threadKey := fmt.Sprintf("thread_invites:%d", threadID) // ← меняем ключ

	inviteIDs, err := r.redisClient.SMembers(ctx, threadKey).Result()
	if err != nil {
		return nil, err
	}

	pipe := r.redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(inviteIDs))

	for i, inviteID := range inviteIDs {
		key := "invite_link:" + inviteID // ← меняем ключ
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var invites []*domain.InviteLink
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}

		var invite domain.InviteLink
		if json.Unmarshal(data, &invite) == nil {
			invites = append(invites, &invite)
		}
	}

	return invites, nil
}

// Удалить invite
func (r *inviteLink) DeleteLink(ctx context.Context, inviteID string) error {
	invite, err := r.GetInviteByID(ctx, inviteID)
	if err != nil {
		return err
	}

	pipe := r.redisClient.TxPipeline()

	// Удаляем основную запись
	pipe.Del(ctx, "invite_link:"+inviteID) // ← меняем ключ

	// Удаляем из индекса
	threadKey := fmt.Sprintf("thread_invites:%d", invite.ThreadID) // ← меняем ключ
	pipe.SRem(ctx, threadKey, inviteID)

	_, err = pipe.Exec(ctx)
	return err
}
