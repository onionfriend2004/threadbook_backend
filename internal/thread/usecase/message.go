package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
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
	// TODO: подумать, как лучше эту структуру впихнуть сюда
	msgResp := &dto.MessageResponse{
		ThreadID: input.ThreadID,
		UserID:   input.UserID,
		Content:  input.Content,
	}
	// Рассылаем сообщение всем участникам
	for _, member := range members {
		if err := uc.wsRepo.PublishToUser(ctx, member.UserID, msgResp); err != nil {
			uc.logger.Warn("failed to publish message to user", zap.Uint("userID", member.UserID), zap.Error(err))
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

func (uc *MessageUsecase) GetSubscribeTokens(ctx context.Context, userID uint) (map[string]string, error) {
	// Получаем список доступных тредов
	threadIDs, err := uc.threadRepo.GetAccessibleThreadIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accessible threads: %w", err)
	}

	// Генерируем токены на каналы
	tokens, err := uc.wsRepo.GenerateSubscribeTokens(ctx, userID, threadIDs, uc.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate channel tokens: %w", err)
	}

	return tokens, nil
}

func (uc *MessageUsecase) GetConnectAndSubscribeTokens(ctx context.Context, userID uint) (ConnectAndSubscribeTokens, error) {
	connectToken, err := uc.GetConnectToken(ctx, userID)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	channelTokens, err := uc.GetSubscribeTokens(ctx, userID)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	return ConnectAndSubscribeTokens{
		ConnectToken:  connectToken,
		ChannelTokens: channelTokens,
	}, nil
}
