package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	spoolExternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type MessageUsecase struct {
	msgRepo    external.MessageRepoInterface
	wsRepo     external.WebsocketRepoInterface
	threadRepo external.ThreadRepoInterface
	spoolRepo  spoolExternal.SpoolRepoInterface
	tokenTTL   time.Duration
	logger     *zap.Logger
}

func NewMessageUsecase(
	msgRepo external.MessageRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	threadRepo external.ThreadRepoInterface,
	spoolRepo spoolExternal.SpoolRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger) *MessageUsecase {
	return &MessageUsecase{
		msgRepo:    msgRepo,
		wsRepo:     wsRepo,
		threadRepo: threadRepo,
		spoolRepo:  spoolRepo,
		tokenTTL:   tokenTTL,
		logger:     logger,
	}
}

func (uc *MessageUsecase) SendMessage(ctx context.Context, input SendMessageInput) (*gdomain.Message, error) {
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrFailedToGetThread
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := uc.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return nil, ErrFailedToGetThread
		}
		if !isMember {
			return nil, ErrThreadNotFound
		}
	} else {
		userStatus, err := uc.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return nil, ErrFailedToGetThread
		}
		if userStatus.IsDeleted || userStatus.AccessLevel < thread.AccessLevel {
			return nil, ErrThreadNotFound
		}
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

	members, err := uc.threadRepo.GetUsersWithAccess(ctx, *thread.SpoolID, thread.AccessLevel)
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
				zap.Uint("userID", member.ID),
				zap.Error(err))
		}
	}

	return msg, nil
}

func (uc *MessageUsecase) GetMessages(ctx context.Context, input GetMessagesInput) ([]gdomain.Message, error) {
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, err
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		uc.threadRepo.IsUserThreadMember(ctx, input.UserID, input.ThreadID)
	}

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
