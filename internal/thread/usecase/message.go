package usecase

import (
	"context"
	"errors"
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
	// Проверяем права пользователя на тред
	hasRights, err := uc.threadRepo.CheckRightsUserOnThreadRoom(ctx, input.ThreadID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check rights: %w", err)
	}
	if !hasRights {
		return nil, errors.New("user has no access to this thread")
	}

	// Проверяем, что тред не закрыт
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if thread.IsClosed {
		return nil, errors.New("cannot send message: thread is closed")
	}

	// Создаём сообщение
	msg := &gdomain.Message{
		ThreadID: input.ThreadID,
		UserID:   input.UserID,
		Content:  input.Content,
		Payloads: input.Payloads,
	}

	// Сохраняем сообщение
	if err := uc.msgRepo.CreateWithPayloads(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Получаем участников треда
	members, err := uc.threadRepo.GetThreadMembers(ctx, input.ThreadID)
	if err != nil {
		return msg, fmt.Errorf("failed to get thread members: %w", err)
	}

	// Готовим событие
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

	// Рассылаем событие всем участникам
	for _, member := range members {
		if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
			uc.logger.Warn("failed to publish message event to user",
				zap.Uint("userID", member.UserID),
				zap.Error(err))
		}
	}

	return msg, nil
}

func (uc *MessageUsecase) GetMessages(ctx context.Context, input GetMessagesInput) ([]gdomain.Message, error) {
	msgs, err := uc.msgRepo.GetByThreadID(ctx, input.ThreadID, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	return msgs, nil
}

func (uc *MessageUsecase) GetConnectToken(ctx context.Context, userID uint) (string, error) {
	token, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate connect token: %w", err)
	}
	return token, nil
}

func (uc *MessageUsecase) GetUserOnlyTokens(ctx context.Context, userID uint) (ConnectAndSubscribeTokens, error) {
	connectToken, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	userChannel := fmt.Sprintf("user#%d", userID)
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

func (uc *MessageUsecase) GetTokensBySpool(ctx context.Context, userID, spoolID uint) (ConnectAndSubscribeTokens, error) {
	threads, err := uc.threadRepo.GetAccessibleThreadIDsBySpool(ctx, userID, spoolID)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	channels := make(map[string]string)
	userChannel := fmt.Sprintf("user#%d", userID)

	connectToken, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	// user channel
	userSub, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, userChannel, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}
	channels[userChannel] = userSub

	// thread channels in this spool
	for _, id := range threads {
		channel := fmt.Sprintf("thread#%d", id)
		token, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, channel, uc.tokenTTL)
		if err != nil {
			uc.logger.Warn("failed gen spool thread sub token", zap.Uint("threadID", id), zap.Error(err))
			continue
		}
		channels[channel] = token
	}

	return ConnectAndSubscribeTokens{
		ConnectToken:  connectToken,
		ChannelTokens: channels,
	}, nil
}
