package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type MessageUsecase struct {
	msgRepo    external.MessageRepoInterface
	wsRepo     external.WebsocketRepoInterface
	threadRepo external.ThreadRepoInterface
	tokenTTL   time.Duration
	logger     *zap.Logger
}

func NewMessageUsecase(
	msgRepo external.MessageRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	threadRepo external.ThreadRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger) *MessageUsecase {
	return &MessageUsecase{
		msgRepo:    msgRepo,
		wsRepo:     wsRepo,
		threadRepo: threadRepo,
		tokenTTL:   tokenTTL,
		logger:     logger,
	}
}

func (uc *MessageUsecase) SendMessage(ctx context.Context, input SendMessageInput) (*gdomain.Message, error) {
	hasRights, err := uc.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil {
		return nil, err
	}
	if !hasRights {
		return nil, ErrNoAccessToThread
	}

	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrFailedToGetThread
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}

	msg := &gdomain.Message{
		ThreadID: input.ThreadID,
		UserID:   input.UserID,
		Content:  input.Content,
		Payloads: input.Payloads,
	}

	if err := uc.msgRepo.CreateWithPayloads(ctx, msg); err != nil {
		return nil, ErrFailedToSaveMsg
	}

	members, err := uc.threadRepo.GetThreadMembers(ctx, input.ThreadID)
	if err != nil {
		return msg, err
	}

	ev := event.Event{
		Type: event.MessageCreated,
		Payload: event.MessageCreatedPayload{
			MessageID: msg.ID,
			ThreadID:  input.ThreadID,
			Content:   input.Content,
			Username:  input.Username,
			CreatedAt: time.Now().Unix(),
		},
	}

	for _, member := range members {
		if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
			uc.logger.Warn(ErrFailedToPublish.Error(),
				zap.Uint("userID", member.UserID),
				zap.Error(err))
		}
	}

	return msg, nil
}

func (uc *MessageUsecase) GetMessages(ctx context.Context, input GetMessagesInput) ([]gdomain.Message, error) {
	return uc.msgRepo.GetByThreadID(ctx, input.ThreadID, input.Limit, input.Offset)
}

func (uc *MessageUsecase) GetConnectToken(ctx context.Context, userID uint) (string, error) {
	return uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
}

func (uc *MessageUsecase) GetUserOnlyTokens(ctx context.Context, userID uint) (ConnectAndSubscribeTokens, error) {
	connectToken, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	userChannel := "user#" + fmt.Sprint(userID)
	subToken, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, userChannel, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	return ConnectAndSubscribeTokens{
		ConnectToken: connectToken,
		ChannelTokens: map[string]string{
			userChannel: subToken,
		},
	}, nil
}

func (uc *MessageUsecase) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*gdomain.Message, error) {
	// Проверяем права пользователя на поток
	hasRights, err := uc.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil {
		return nil, err
	}
	if !hasRights {
		return nil, ErrNoAccessToThread
	}

	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrFailedToGetThread
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}

	// Получаем сообщение
	msg, err := uc.msgRepo.GetByID(ctx, input.MessageID)
	if err != nil {
		return nil, err
	}

	// Проверка, что сообщение принадлежит потоку
	if msg.ThreadID != input.ThreadID {
		return nil, ErrMessageNotFound
	}

	// Проверка, что пользователь — автор сообщения
	if msg.UserID != input.UserID {
		return nil, ErrNoAccessToMessage
	}

	// Обновляем контент
	msg.Content = input.Content
	if err := uc.msgRepo.Update(ctx, msg); err != nil {
		return nil, err
	}

	// Публикуем событие через WS
	ev := event.Event{
		Type: event.MessageUpdated,
		Payload: event.MessageUpdatedPayload{
			MessageID: msg.ID,
			ThreadID:  msg.ThreadID,
			Content:   msg.Content,
			UpdatedAt: msg.UpdatedAt.Unix(),
		},
	}

	members, err := uc.threadRepo.GetThreadMembers(ctx, input.ThreadID)
	if err == nil {
		for _, member := range members {
			if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
				uc.logger.Warn("failed to publish updated message",
					zap.Uint("userID", member.UserID),
					zap.Error(err))
			}
		}
	}

	return msg, nil
}

func (uc *MessageUsecase) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	// Проверяем права пользователя на поток
	hasRights, err := uc.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil {
		return err
	}
	if !hasRights {
		return ErrNoAccessToThread
	}

	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return ErrFailedToGetThread
	}
	if thread.IsClosed {
		return ErrThreadIsClosed
	}

	// Получаем сообщение
	msg, err := uc.msgRepo.GetByID(ctx, input.MessageID)
	if err != nil {
		return err
	}

	// Проверка, что сообщение принадлежит потоку
	if msg.ThreadID != input.ThreadID {
		return ErrMessageNotFound
	}

	// Проверка, что пользователь — автор сообщения
	if msg.UserID != input.UserID {
		return ErrNoAccessToMessage
	}

	// Удаляем сообщение
	if err := uc.msgRepo.DeleteByID(ctx, input.MessageID); err != nil {
		return err
	}

	// Публикуем событие через WS
	ev := event.Event{
		Type: event.MessageDeleted,
		Payload: event.MessageDeletedPayload{
			MessageID: msg.ID,
			ThreadID:  msg.ThreadID,
		},
	}

	members, err := uc.threadRepo.GetThreadMembers(ctx, input.ThreadID)
	if err == nil {
		for _, member := range members {
			if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
				uc.logger.Warn("failed to publish deleted message",
					zap.Uint("userID", member.UserID),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (uc *MessageUsecase) GetTokensBySpool(ctx context.Context, userID, spoolID uint) (ConnectAndSubscribeTokens, error) {
	threads, err := uc.threadRepo.GetAccessibleThreadIDsBySpool(ctx, userID, spoolID)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	channels := make(map[string]string)
	userChannel := "user#" + fmt.Sprint(userID)

	connectToken, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	userSub, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, userChannel, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}
	channels[userChannel] = userSub

	for _, id := range threads {
		channel := "thread#" + fmt.Sprint(id)
		token, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, channel, uc.tokenTTL)
		if err != nil {
			uc.logger.Warn(ErrFailedToPublish.Error(),
				zap.Uint("threadID", id),
				zap.Error(err))
			continue
		}
		channels[channel] = token
	}

	return ConnectAndSubscribeTokens{
		ConnectToken:  connectToken,
		ChannelTokens: channels,
	}, nil
}
