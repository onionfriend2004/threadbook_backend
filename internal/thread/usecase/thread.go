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
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error)
	CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, input InviteToThreadInput) error
	UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error)
}

type ThreadUsecase struct {
	threadRepo external.ThreadRepoInterface
	wsRepo     external.WebsocketRepoInterface
	userRepo   userexternal.UserRepoInterface
	tokenTTL   time.Duration
	logger     *zap.Logger
}

func NewThreadUsecase(
	threadRepo external.ThreadRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	userRepo userexternal.UserRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo: threadRepo,
		wsRepo:     wsRepo,
		userRepo:   userRepo,
		tokenTTL:   tokenTTL,
		logger:     logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error) {
	if !(input.TypeThread == "private" || input.TypeThread == "public") {
		return nil, ErrWrognTypeThread
	}

	newThread, err := u.threadRepo.Create(ctx, input.OwnerID, input.SpoolID, input.Title, input.TypeThread)
	if err != nil {
		return nil, err
	}

	threadChannel := fmt.Sprintf("thread#%d", newThread.ID)

	subToken, err := u.wsRepo.GenerateSubscribeToken(ctx, input.OwnerID, threadChannel, u.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate subscribe token: %w", err)
	}

	members, err := u.threadRepo.GetThreadMembers(ctx, newThread.ID)
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
