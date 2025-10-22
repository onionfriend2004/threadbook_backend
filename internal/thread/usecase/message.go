package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
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

// Конструктор
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

// ---------- Отправка сообщения ----------
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

	// Рассылаем сообщение всем участникам
	for _, member := range members {
		if err := uc.wsRepo.PublishToUser(ctx, member.UserID, msg); err != nil {
			// не прерываем рассылку, просто логируем ошибку
			fmt.Printf("warn: failed to publish message to user %d: %v\n", member.UserID, err)
		}
	}

	return msg, nil
}

// ---------- Получение сообщений треда ----------
func (uc *MessageUsecase) GetMessages(ctx context.Context, input GetMessagesInput) ([]gdomain.Message, error) {
	msgs, err := uc.msgRepo.GetByThreadID(ctx, input.ThreadID, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	return msgs, nil
}

// ---------- Получение токена для подключения к WS ----------
func (uc *MessageUsecase) GetSubscribeToken(ctx context.Context, input GetSubscribeTokenInput) (string, error) {
	// теперь токен выдаётся не на тред, а на глобальный канал пользователя
	token, err := uc.wsRepo.GenerateUserToken(ctx, input.UserID, uc.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate user subscribe token: %w", err)
	}
	return token, nil
}
