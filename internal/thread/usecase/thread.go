package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	userexternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external/domain"
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error)
	CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, input InviteToThreadInput) error
	UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error)
	CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*domain.InviteLink, error)
	JoinToThread(ctx context.Context, input JoinToThreadInput) error
	DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error
	GetThreadInviteLinks(ctx context.Context, input GetThreadInviteLinksInput) ([]*domain.InviteLink, error)
}

type ThreadUsecase struct {
	threadRepo external.ThreadRepoInterface
	noAuthRepo repo.InviteLinkRepoInterface
	wsRepo     external.WebsocketRepoInterface
	userRepo   userexternal.UserRepoInterface
	tokenTTL   time.Duration
	logger     *zap.Logger
}

func NewThreadUsecase(
	threadRepo external.ThreadRepoInterface,
	noAuthRepo repo.InviteLinkRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	userRepo userexternal.UserRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo: threadRepo,
		noAuthRepo: noAuthRepo,
		wsRepo:     wsRepo,
		userRepo:   userRepo,
		tokenTTL:   tokenTTL,
		logger:     logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error) {
	if !(input.TypeThread == "private" || input.TypeThread == "public") {
		return nil, ErrWrongTypeThread
	}

	// Создаём поток в репозитории
	newThread, err := u.threadRepo.Create(ctx, input.OwnerID, input.SpoolID, input.Title, input.TypeThread)
	if err != nil {
		return nil, err
	}

	threadChannel := fmt.Sprintf("thread#%d", newThread.ID)

	// Получаем список участников потока
	members, err := u.threadRepo.GetThreadMembers(ctx, newThread.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread members: %w", err)
	}

	// Для каждого участника создаём свой токен и рассылаем событие
	for _, member := range members {
		subToken, err := u.wsRepo.GenerateSubscribeToken(ctx, member.UserID, threadChannel, u.tokenTTL)
		if err != nil {
			u.logger.Warn("failed to generate subscribe token", zap.Uint("userID", member.UserID), zap.Error(err))
			continue
		}

		eventPayload := event.ThreadCreatedPayload{
			ThreadID:  newThread.ID,
			SpoolID:   newThread.SpoolID,
			Title:     newThread.Title,
			CreatedAt: newThread.CreatedAt.Unix(),
			Channel:   threadChannel,
			Token:     subToken,
		}

		if err := u.wsRepo.PublishToUser(ctx, member.UserID, event.Event{
			Type:    event.ThreadCreated,
			Payload: eventPayload,
		}); err != nil {
			u.logger.Warn("failed to publish thread created event", zap.Uint("userID", member.UserID), zap.Error(err))
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
	thread, err := u.threadRepo.CloseThread(input.ThreadID, input.UserID)
	if err != nil {
		return nil, err
	}

	// Получаем участников треда
	members, err := u.threadRepo.GetThreadMembers(ctx, thread.ID)
	if err != nil {
		u.logger.Warn("failed to get thread members for CloseThread event", zap.Error(err))
		return thread, nil // возвращаем закрытый тред даже если событие не отправилось
	}

	// Подготавливаем payload события
	payload := event.ThreadClosedPayload{
		ThreadID: thread.ID,
		SpoolID:  thread.SpoolID,
	}

	// Рассылаем событие всем участникам
	for _, member := range members {
		if err := u.wsRepo.PublishToUser(ctx, member.UserID, event.Event{
			Type:    event.ThreadDeleted,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish ThreadDeleted event", zap.Uint("userID", member.UserID), zap.Error(err))
		}
	}

	return thread, nil
}

func (u *ThreadUsecase) InviteToThread(ctx context.Context, input InviteToThreadInput) error {
	// Добавляем пользователей в тред через репозиторий
	if err := u.threadRepo.InviteToThread(ctx, input.InviterID, input.InviteeUsernames, input.ThreadID); err != nil {
		return err
	}

	// Получаем сам тред
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	threadChannel := fmt.Sprintf("thread#%d", thread.ID)

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

		payload := event.ThreadInvitePayload{
			ThreadID: thread.ID,
			SpoolID:  thread.SpoolID,
			Type:     thread.Type,
			Title:    thread.Title,
			Channel:  threadChannel,
			Token:    subToken,
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

	updatedThread, err := u.threadRepo.Update(ctx, input.ID, input.EditorID, input.Title, input.ThreadType)
	if err != nil {
		return nil, err
	}

	// Получаем участников треда
	members, err := u.threadRepo.GetThreadMembers(ctx, updatedThread.ID)
	if err != nil {
		u.logger.Warn("failed to get thread members for ThreadUpdated event", zap.Error(err))
		return updatedThread, nil
	}

	// Подготавливаем payload события
	payload := event.ThreadUpdatedPayload{
		ThreadID:  updatedThread.ID,
		SpoolID:   updatedThread.SpoolID,
		Title:     updatedThread.Title,
		UpdatedAt: updatedThread.UpdatedAt.Unix(),
	}

	// Рассылаем событие всем участникам
	for _, member := range members {
		if err := u.wsRepo.PublishToUser(ctx, member.UserID, event.Event{
			Type:    event.ThreadUpdated,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish ThreadUpdated event", zap.Uint("userID", member.UserID), zap.Error(err))
		}
	}

	return updatedThread, nil
}

func (u *ThreadUsecase) CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*domain.InviteLink, error) {

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

	session, err := u.noAuthRepo.CreateLink(ctx, input.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to create noauth session: %w", err)
	}

	return session, nil
}

func (u *ThreadUsecase) JoinToThread(ctx context.Context, input JoinToThreadInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	session, err := u.noAuthRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrThreadNotFound
	}

	return u.threadRepo.AddUserToThread(ctx, input.UserID, session.ThreadID)
}

func (u *ThreadUsecase) DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	session, err := u.noAuthRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrThreadNotFound
	}

	isOwner, err := u.threadRepo.IsThreadOwner(ctx, input.UserID, session.ThreadID)
	if err != nil || !isOwner {
		return ErrThreadNotFound
	}

	return u.noAuthRepo.DeleteLink(ctx, session.ID)
}

func (u *ThreadUsecase) GetThreadInviteLinks(ctx context.Context, input GetThreadInviteLinksInput) ([]*domain.InviteLink, error) {
	inThread, err := u.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil || !inThread {
		return nil, ErrThreadNotFound
	}

	return u.noAuthRepo.GetLinksByThread(ctx, input.ThreadID)
}
