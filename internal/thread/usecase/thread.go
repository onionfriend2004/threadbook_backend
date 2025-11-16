package usecase

import (
	"context"
	"fmt"
	"time"

	userexternal "github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"

	spoolexternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
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
	GetThreadUsers(ctx context.Context, input GetThreadUsersInput) ([]gdomain.User, error)
}

type ThreadUsecase struct {
	threadRepo     external.ThreadRepoInterface
	spoolRepo      spoolexternal.SpoolRepoInterface
	InviteLinkRepo repo.InviteLinkRepoInterface
	wsRepo         external.WebsocketRepoInterface
	userRepo       userexternal.UserRepoInterface
	tokenTTL       time.Duration
	logger         *zap.Logger
}

func NewThreadUsecase(
	threadRepo external.ThreadRepoInterface,
	spoolRepo spoolexternal.SpoolRepoInterface,
	InviteLinkRepo repo.InviteLinkRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	userRepo userexternal.UserRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo:     threadRepo,
		spoolRepo:      spoolRepo,
		InviteLinkRepo: InviteLinkRepo,
		wsRepo:         wsRepo,
		userRepo:       userRepo,
		tokenTTL:       tokenTTL,
		logger:         logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error) {
	if input.SpoolID != nil {
		userSpool, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.OwnerID, *input.SpoolID)
		if err != nil {
			return nil, external.ErrUserNotInSpool
		}
		var trueAccessLevel uint
		if input.AccessLevel == nil {
			trueAccessLevel = 0
		} else {
			trueAccessLevel = *input.AccessLevel
		}
		// Если объединить в один if -- потом ничего непонятно
		if userSpool.IsDeleted {
			return nil, external.ErrPermissionDenied
		}
		if input.ThreadType == gdomain.ThreadTypePrivate && userSpool.AccessLevel < 1 {
			return nil, external.ErrPermissionDenied
		}
		if input.ThreadType == gdomain.ThreadTypePublic && (userSpool.AccessLevel <= trueAccessLevel) {
			return nil, external.ErrPermissionDenied
		}
	} else {
		if input.ThreadType == gdomain.ThreadTypePublic {
			return nil, ErrInvalidInput
		}
	}

	newThread, err := u.threadRepo.Create(ctx, input.OwnerID, input.SpoolID, input.Title, input.ThreadType, input.AccessLevel)
	if err != nil {
		return nil, err
	}
	if newThread.Type == gdomain.ThreadTypePublic {
		threadChannel := fmt.Sprintf("thread#%d", newThread.ID)

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
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, err
	}
	var members []gdomain.User
	if thread.Type == gdomain.ThreadTypePrivate {
		isOwner, err := u.threadRepo.IsThreadOwner(ctx, input.UserID, input.ThreadID)
		// isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, input.ThreadID)
		if err != nil {
			return nil, err
		}
		if isOwner {
			thread, err = u.threadRepo.CloseThread(ctx, thread.ID)
			if err != nil {
				return nil, err
			}
		}
		members, err = u.threadRepo.GetThreadUsers(ctx, thread.ID)
		if err != nil {
			u.logger.Warn("failed to get thread members for CloseThread event", zap.Error(err))
			return thread, nil // возвращаем закрытый тред даже если событие не отправилось
		}
	} else if thread.SpoolID != nil {
		user, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return nil, err
		}
		if user.AccessLevel <= thread.AccessLevel {
			return nil, ErrNoAccessToThread
		}
		thread, err = u.threadRepo.CloseThread(ctx, thread.ID)
		if err != nil {
			return nil, err
		}
		members, err = u.threadRepo.GetUsersWithAccess(ctx, *thread.SpoolID, thread.AccessLevel)
		if err != nil {
			u.logger.Warn("failed to get thread members for CloseThread event", zap.Error(err))
			return thread, nil // возвращаем закрытый тред даже если событие не отправилось
		}
	} else {
		return nil, ErrInvalidInput
	}

	// Подготавливаем payload события
	payload := event.ThreadClosedPayload{
		ThreadID: thread.ID,
		SpoolID:  thread.SpoolID,
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
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return err
	}
	if thread.Type == gdomain.ThreadTypePublic {
		return ErrWrongTypeThread
	}
	isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.InviterID, input.ThreadID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrThreadNotFound
	}
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
	// userAccessLevel, threadAccessLevel, err := u.threadRepo.GetUserThreadAccessLevelsToUpdate(ctx, input.EditorID, input.ID)
	// if err != nil {
	// 	return nil, ErrFailToGetThread
	// }

	// // Проверяем доступ к редактированию
	// if userAccessLevel <= threadAccessLevel {
	// 	return nil, ErrNoAccessToThread
	// }

	// // Если передали новый access level, проверяем что у пользователя достаточно прав
	// if input.AccessLevel != nil && userAccessLevel <= *input.AccessLevel {
	// 	return nil, ErrNoAccessToThread
	// }
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, err
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isOwner, err := u.threadRepo.IsThreadOwner(ctx, input.EditorID, thread.ID)
		// isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.EditorID, thread.ID)
		if err != nil {
			return nil, err
		}
		if !isOwner {
			return nil, ErrThreadNotFound
		}
	} else {
		userStatus, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.EditorID, *thread.SpoolID)
		if err != nil {
			return nil, err
		}
		if userStatus.AccessLevel <= thread.AccessLevel {
			return nil, ErrThreadNotFound
		}
	}

	updatedThread, err := u.threadRepo.Update(ctx, input.ThreadID, input.Title, input.ThreadType, input.AccessLevel)
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
		SpoolID:   updatedThread.SpoolID,
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
		return nil, err
	}
	if thread.Type != gdomain.ThreadTypePrivate {
		return nil, ErrWrongTypeThread
	}
	if thread.IsClosed {
		return nil, ErrThreadClosed
	}
	isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
	if err != nil {
		return nil, ErrThreadNotFound
	}
	if !isMember {
		return nil, ErrThreadNotFound
	}

	session, err := u.InviteLinkRepo.CreateLink(ctx, "thread", input.ThreadID, input.UserID, input.ExpiresAt, input.MaxUses)
	if err != nil {
		return nil, fmt.Errorf("failed to create noauth session: %w", err)
	}

	return session, nil
}

func (u *ThreadUsecase) JoinToThread(ctx context.Context, input JoinToThreadInput) error {
	// if input.InviteToken == "" {
	// 	return ErrInvalidInput
	// }

	session, err := u.InviteLinkRepo.GetInviteByID(ctx, input.InviteToken)
	if err != nil {
		return ErrThreadNotFound
	}
	thread, err := u.threadRepo.GetThreadByID(ctx, session.ResourceID)
	if err != nil {
		return err
	}
	if thread.SpoolID != nil {
		user, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return err
		}
		if user == nil {
			u.spoolRepo.AddUserToSpoolByUsername(ctx, input.Username, *thread.SpoolID)
		}
	}
	return u.threadRepo.AddUserToThread(ctx, input.UserID, session.ResourceID)
}

func (u *ThreadUsecase) DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error {
	session, err := u.InviteLinkRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrThreadNotFound
	}

	thread, err := u.threadRepo.GetThreadByID(ctx, session.ResourceID)
	if err != nil {
		return err
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return err
		}
		if !isMember {
			return ErrThreadNotFound
		}
	} else {
		return ErrWrongTypeThread
	}

	// user, err := u.threadRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
	// if err != nil || user == nil {
	// 	return err
	// }

	// isOwner, err := u.threadRepo.IsThreadOwner(ctx, input.UserID, session.ResourceID)
	// if err != nil || !isOwner {
	// 	return ErrThreadNotFound
	// }

	return u.InviteLinkRepo.DeleteLink(ctx, session.ID)
}

func (u *ThreadUsecase) GetThreadInviteLinks(ctx context.Context, input GetThreadInviteLinksInput) ([]*gdomain.InviteLink, error) {
	fmt.Printf("GetThreadInviteLinks input: ThreadID=%d, UserID=%d\n", input.ThreadID, input.UserID)
	isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrThreadNotFound
	}

	return u.InviteLinkRepo.GetLinksByResource(ctx, "thread", input.ThreadID)
}

func (u *ThreadUsecase) GetThreadUsers(ctx context.Context, input GetThreadUsersInput) ([]gdomain.User, error) {
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, err
	}
	if thread.Type != gdomain.ThreadTypePrivate {
		return nil, ErrWrongTypeThread
	}
	isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrThreadNotFound
	}
	users, err := u.threadRepo.GetThreadUsers(ctx, input.ThreadID)
	if err != nil {
		return nil, err
	}
	return users, err
}
