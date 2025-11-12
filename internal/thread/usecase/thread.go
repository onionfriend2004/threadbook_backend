package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	userexternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"

	// spoolexternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error)
	CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, input InviteToThreadInput) error
	UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error)
	CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*gdomain.InviteLink, error)
	JoinToThread(ctx context.Context, input JoinToThreadInput) error
	DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error
	GetThreadInviteLinks(ctx context.Context, input GetThreadInviteLinksInput) ([]*gdomain.InviteLink, error)
}

type ThreadUsecase struct {
	threadRepo     external.ThreadRepoInterface
	InviteLinkRepo repo.InviteLinkRepoInterface
	wsRepo         external.WebsocketRepoInterface
	userRepo       userexternal.UserRepoInterface
	tokenTTL       time.Duration
	logger         *zap.Logger
}

func NewThreadUsecase(
	threadRepo external.ThreadRepoInterface,
	InviteLinkRepo repo.InviteLinkRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	userRepo userexternal.UserRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo:     threadRepo,
		InviteLinkRepo: InviteLinkRepo,
		wsRepo:         wsRepo,
		userRepo:       userRepo,
		tokenTTL:       tokenTTL,
		logger:         logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error) {
	if !(input.TypeThread == "private" || input.TypeThread == "public") {
		return nil, ErrWrongTypeThread
	}

	if input.SpoolID != nil {
		userSpool, err := u.threadRepo.GetUserSpoolStatus(ctx, input.OwnerID, *input.SpoolID)
		if err != nil {
			return nil, external.ErrUserNotInSpool
		}
		if input.TypeThread != "private" && (userSpool.AccessLevel <= input.AccessLevel || userSpool.IsDeleted) {
			return nil, external.ErrPermissionDenied
		}
	}

	newThread, err := u.threadRepo.Create(ctx, input.OwnerID, input.SpoolID, input.Title, input.TypeThread, input.AccessLevel)
	if err != nil {
		return nil, err
	}
	if input.SpoolID != nil {
		threadChannel := fmt.Sprintf("thread#%d", newThread.ID)

		subToken, err := u.wsRepo.GenerateSubscribeToken(ctx, input.OwnerID, threadChannel, u.tokenTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate subscribe token: %w", err)
		}

		members, err := u.threadRepo.GetUsersWithAccess(ctx, *input.SpoolID, input.AccessLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to get thread members: %w", err)
		}

		eventPayload := event.ThreadCreatedPayload{
			ThreadID:       newThread.ID,
			Title:          newThread.Title,
			CreatedAt:      newThread.CreatedAt.Unix(),
			Channel:        threadChannel,
			Token:          subToken,
			SubscribeToken: subToken,
		}

		for _, member := range members {
			if err := u.wsRepo.PublishToUser(ctx, member.ID, event.Event{
				Type:    event.ThreadCreated,
				Payload: eventPayload,
			}); err != nil {
				u.logger.Warn("failed to publish thread created event", zap.Uint("userID", member.ID), zap.Error(err))
			}
		}
	}
	return newThread, nil
}

func (u *ThreadUsecase) GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error) {
	newThread, err := u.threadRepo.GetBySpoolID(ctx, input.UserID, input.SpoolID)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error) {
	userAccessLevel, threadAccessLevel, err := u.threadRepo.GetUserThreadAccessLevelsToUpdate(ctx, input.ThreadID, input.UserID)
	if err != nil {
		return nil, err
	}
	if userAccessLevel <= threadAccessLevel {
		return nil, ErrNoAccessToThread
	}

	thread, err := u.threadRepo.CloseThread(input.ThreadID, input.UserID)
	if err != nil {
		return nil, err
	}

	// Получаем участников треда
	members, err := u.threadRepo.GetUsersWithAccess(ctx, *thread.SpoolID, thread.AccessLevel)
	if err != nil {
		u.logger.Warn("failed to get thread members for CloseThread event", zap.Error(err))
		return thread, nil // возвращаем закрытый тред даже если событие не отправилось
	}

	// Подготавливаем payload события
	payload := event.ThreadClosedPayload{
		ThreadID: thread.ID,
	}

	// Рассылаем событие всем участникам
	for _, member := range members {
		if err := u.wsRepo.PublishToUser(ctx, member.ID, event.Event{
			Type:    event.ThreadDeleted,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish ThreadDeleted event", zap.Uint("userID", member.ID), zap.Error(err))
		}
	}

	return thread, nil
}

func (u *ThreadUsecase) InviteToThread(ctx context.Context, input InviteToThreadInput) error {
	// Добавляем пользователей в тред через репозиторий
	if err := u.threadRepo.InviteToThread(ctx, input.InviterID, input.InviteeUsernames, input.ThreadID); err != nil {
		return err
	}

	threadChannel := fmt.Sprintf("thread#%d", input.ThreadID)

	for _, username := range input.InviteeUsernames {
		user, err := u.userRepo.GetUserByUsername(ctx, username)
		if err != nil {
			u.logger.Warn("failed to get user ID by username", zap.String("username", username), zap.Error(err))
			continue // не блокируем остальных пользователей
		}

		subToken, err := u.wsRepo.GenerateSubscribeToken(ctx, user.ID, threadChannel, u.tokenTTL)
		if err != nil {
			u.logger.Warn("failed to generate subscribe token for invited user", zap.String("username", username), zap.Error(err))
			continue
		}

		payload := event.ThreadSubTokenPayload{
			Channel: threadChannel,
			Token:   subToken,
		}

		if err := u.wsRepo.PublishToUser(ctx, user.ID, event.Event{
			Type:    event.ThreadInvited,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish ThreadInvited event", zap.String("username", username), zap.Error(err))
		}
	}

	return nil
}

func (u *ThreadUsecase) UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error) {
	if input.ID == 0 {
		return nil, errors.New("thread id is required")
	}
	if input.EditorID == 0 {
		return nil, errors.New("editor id is required")
	}

	userAccessLevel, threadAccessLevel, err := u.threadRepo.GetUserThreadAccessLevelsToUpdate(ctx, input.EditorID, input.ID)
	if err != nil {
		return nil, ErrFailToGetThread
	}
	if userAccessLevel <= threadAccessLevel || userAccessLevel <= *input.AccessLevel {
		return nil, ErrNoAccessToThread
	}

	updatedThread, err := u.threadRepo.Update(ctx, input.ID, input.Title, input.ThreadType, input.AccessLevel)
	if err != nil {
		return nil, err
	}

	// Получаем участников треда
	if updatedThread.SpoolID != nil {
		members, err := u.threadRepo.GetUsersWithAccess(ctx, *updatedThread.SpoolID, updatedThread.AccessLevel)
		if err != nil {
			u.logger.Warn("failed to get thread members for ThreadUpdated event", zap.Error(err))
			return updatedThread, nil
		}

		// Подготавливаем payload события
		payload := event.ThreadUpdatedPayload{
			ThreadID:  updatedThread.ID,
			Title:     updatedThread.Title,
			UpdatedAt: updatedThread.UpdatedAt.Unix(),
		}

		// Рассылаем событие всем участникам
		for _, member := range members {
			if err := u.wsRepo.PublishToUser(ctx, member.ID, event.Event{
				Type:    event.ThreadUpdated,
				Payload: payload,
			}); err != nil {
				u.logger.Warn("failed to publish ThreadUpdated event", zap.Uint("userID", member.ID), zap.Error(err))
			}
		}
	}
	return updatedThread, nil
}

func (u *ThreadUsecase) CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*gdomain.InviteLink, error) {

	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if thread.CreatorID != input.UserID {
		return nil, ErrNoAccessToThread
	}
	if thread.IsClosed {
		return nil, ErrThreadClosed
	}

	session, err := u.InviteLinkRepo.CreateLink(ctx, "thread", input.ThreadID, input.UserID, input.ExpiresAt, input.MaxUses)
	if err != nil {
		return nil, fmt.Errorf("failed to create noauth session: %w", err)
	}

	return session, nil
}

func (u *ThreadUsecase) JoinToThread(ctx context.Context, input JoinToThreadInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	session, err := u.InviteLinkRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrThreadNotFound
	}

	return u.threadRepo.AddUserToThread(ctx, input.UserID, session.ResourceID)
}

func (u *ThreadUsecase) DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	session, err := u.InviteLinkRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrThreadNotFound
	}

	isOwner, err := u.threadRepo.IsThreadOwner(ctx, input.UserID, session.ResourceID)
	if err != nil || !isOwner {
		return ErrThreadNotFound
	}

	return u.InviteLinkRepo.DeleteLink(ctx, session.ID)
}

func (u *ThreadUsecase) GetThreadInviteLinks(ctx context.Context, input GetThreadInviteLinksInput) ([]*gdomain.InviteLink, error) {
	inThread, err := u.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil || !inThread {
		return nil, ErrThreadNotFound
	}

	return u.InviteLinkRepo.GetLinksByResource(ctx, "thread", input.ThreadID)
}
