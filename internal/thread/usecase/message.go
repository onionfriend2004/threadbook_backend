package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	spoolExternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type MessageUsecase struct {
	msgRepo    external.MessageRepoInterface
	wsRepo     external.WebsocketRepoInterface
	fileUC     usecase.FileUsecaseInterface
	threadRepo external.ThreadRepoInterface
	spoolRepo  spoolExternal.SpoolRepoInterface
	tokenTTL   time.Duration
	logger     *zap.Logger
}

func NewMessageUsecase(
	msgRepo external.MessageRepoInterface,
	wsRepo external.WebsocketRepoInterface,
	fileUC usecase.FileUsecaseInterface,
	threadRepo external.ThreadRepoInterface,
	spoolRepo spoolExternal.SpoolRepoInterface,
	tokenTTL time.Duration,
	logger *zap.Logger) *MessageUsecase {
	return &MessageUsecase{
		msgRepo:    msgRepo,
		wsRepo:     wsRepo,
		fileUC:     fileUC,
		threadRepo: threadRepo,
		spoolRepo:  spoolRepo,
		tokenTTL:   tokenTTL,
		logger:     logger,
	}
}

func (uc *MessageUsecase) SendMessage(ctx context.Context, input SendMessageInput) (*gdomain.Message, error) {
	// --- 1. Проверяем права пользователя ---
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrThreadNotFound
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := uc.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if !isMember {
			return nil, ErrThreadNotFound
		}
	} else {
		userStatus, err := uc.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return nil, ErrThreadNotFound
		}
	}
	// --- 2. Загружаем файлы (payloads) ---
	var uploadedFiles []string
	filesSaved := false

	for _, payload := range input.Payloads {
		fileInput := usecase.SaveFile{
			File:        payload.File,
			Size:        payload.Size,
			Filename:    payload.Filename,
			ContentType: payload.ContentType,
			UserID:      strconv.FormatUint(uint64(input.UserID), 10),
			FileType:    "message_payload",
		}

		fileLink, err := uc.fileUC.SaveFile(ctx, fileInput)
		if err != nil {
			// rollback всех загруженных файлов
			for _, saved := range uploadedFiles {
				if delErr := uc.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{Filename: saved}); delErr != nil {
					uc.logger.Error("failed to rollback uploaded file", zap.Error(delErr), zap.String("filename", saved))
				}
			}
			return nil, ErrFailedToSaveMessageFiles
		}
		uploadedFiles = append(uploadedFiles, fileLink)
	}
	filesSaved = true

	defer func(uploaded []string, saved bool) {
		if !saved {
			for _, fileLink := range uploaded {
				if delErr := uc.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{Filename: fileLink}); delErr != nil {
					uc.logger.Error("failed to cleanup message file after error", zap.Error(delErr), zap.String("file_link", fileLink))
				}
			}
		}
	}(uploadedFiles, filesSaved)

	// --- 3. Формируем gdomain.Message с payloads ---
	newMsg := &gdomain.Message{
		ThreadID:  input.ThreadID,
		UserID:    input.UserID,
		Content:   input.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, link := range uploadedFiles {
		newMsg.Payloads = append(newMsg.Payloads, gdomain.MessagePayload{
			FileLink:  link,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// --- 4. Сохраняем сообщение в транзакции ---
	err = uc.msgRepo.WithTx(ctx, func(txCtx context.Context) error {
		return uc.msgRepo.CreateWithPayloads(txCtx, newMsg)
	})
	if err != nil {
		uc.logger.Error("failed to create message in database", zap.Error(err), zap.Uint("user_id", input.UserID), zap.Uint("thread_id", input.ThreadID))
		return nil, ErrFailedToSaveMsg
	}

	uc.logger.Info("message created successfully", zap.Uint("message_id", newMsg.ID), zap.Int("payload_count", len(uploadedFiles)))

	// --- 5. Публикуем событие через WebSocket ---
	payloadLinks := make([]string, len(newMsg.Payloads))
	for i, p := range newMsg.Payloads {
		payloadLinks[i] = p.FileLink
	}
	// if thread.Type == gdomain.ThreadTypePublic {

	// }
	// members, err := uc.threadRepo.GetUsersWithAccess(ctx, *thread.SpoolID, thread.AccessLevel)
	// if err != nil {
	// 	return newMsg, err
	// }

	ev := event.Event{
		Type: event.MessageCreated,
		Payload: event.MessageCreatedPayload{
			MessageID: newMsg.ID,
			ThreadID:  newMsg.ThreadID,
			Content:   newMsg.Content,
			Username:  input.Username,
			Payloads:  payloadLinks,
			CreatedAt: newMsg.CreatedAt.Unix(),
		},
	}

	if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
		uc.logger.Warn("failed to publish message event to WS",
			zap.Uint("thread_id", input.ThreadID),
			zap.Error(err))
	}
	// for _, member := range members {
	// if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
	// 	uc.logger.Warn(ErrFailedToPublish.Error(),
	// 		zap.Uint("userID", member.ID),
	// 		zap.Error(err))
	// }
	// }

	return newMsg, nil
}

func (uc *MessageUsecase) GetMessages(ctx context.Context, input GetMessagesInput) ([]gdomain.Message, error) {
	// thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	// if err != nil {
	// 	return nil, err
	// }
	// if thread.Type == gdomain.ThreadTypePrivate {
	// 	uc.threadRepo.IsUserThreadMember(ctx, input.UserID, input.ThreadID)
	// }

	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrThreadNotFound
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := uc.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if !isMember {
			return nil, ErrThreadNotFound
		}
	} else {
		userStatus, err := uc.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return nil, ErrThreadNotFound
		}
	}

	return uc.msgRepo.GetByThreadCursor(ctx, input.ThreadID, input.CursorID, input.Limit, input.Forward)
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
	subToken, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, userChannel)
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
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return nil, ErrThreadNotFound
	}
	if thread.IsClosed {
		return nil, ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := uc.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if !isMember {
			return nil, ErrThreadNotFound
		}
	} else {
		userStatus, err := uc.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return nil, ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return nil, ErrThreadNotFound
		}
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

	if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
		uc.logger.Warn("failed to publish message event to WS",
			zap.Uint("thread_id", input.ThreadID),
			zap.Error(err))
	}

	return msg, nil
}

func (uc *MessageUsecase) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	// Проверка прав
	thread, err := uc.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return ErrThreadNotFound
	}
	if thread.IsClosed {
		return ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := uc.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return ErrThreadNotFound
		}
		if !isMember {
			return ErrThreadNotFound
		}
	} else {
		userStatus, err := uc.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return ErrThreadNotFound
		}
	}
	msg, err := uc.msgRepo.GetByID(ctx, input.MessageID)
	if err != nil {
		return err
	}

	if msg.ThreadID != input.ThreadID {
		return ErrMessageNotFound
	}
	if msg.UserID != input.UserID {
		return ErrNoAccessToMessage
	}

	// Транзакция на удаление сообщения
	err = uc.msgRepo.WithTx(ctx, func(txCtx context.Context) error {
		// Удаляем сообщение в БД
		if err := uc.msgRepo.DeleteByID(txCtx, input.MessageID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Удаляем payload файлы уже после успешного удаления записи
	for _, p := range msg.Payloads {
		if delErr := uc.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{
			Filename: p.FileLink,
		}); delErr != nil {
			uc.logger.Warn("failed to delete file from MinIO",
				zap.String("filename", p.FileLink),
				zap.Error(delErr))
		}
	}

	// WS событие
	ev := event.Event{
		Type: event.MessageDeleted,
		Payload: event.MessageDeletedPayload{
			MessageID: msg.ID,
			ThreadID:  msg.ThreadID,
		},
	}

	if err := uc.wsRepo.PublishToThread(ctx, input.ThreadID, ev); err != nil {
		uc.logger.Warn("failed to publish message event to WS",
			zap.Uint("thread_id", input.ThreadID),
			zap.Error(err))
	}

	return nil
}

func (uc *MessageUsecase) GetTokensBySpool(ctx context.Context, userID, spoolID uint) (ConnectAndSubscribeTokens, error) {
	status, err := uc.spoolRepo.GetUserSpoolStatus(ctx, userID, spoolID)
	if err != nil || status == nil {
		return ConnectAndSubscribeTokens{}, err
	}
	threads, err := uc.spoolRepo.GetSpoolThreadsByUser(ctx, *status)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	channels := make(map[string]string)
	userChannel := "user#" + fmt.Sprint(userID)

	connectToken, err := uc.wsRepo.GenerateConnectToken(ctx, userID, uc.tokenTTL)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}

	userSub, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, userChannel)
	if err != nil {
		return ConnectAndSubscribeTokens{}, err
	}
	channels[userChannel] = userSub

	for _, id := range threads {
		channel := "thread#" + fmt.Sprint(id.ID)
		token, err := uc.wsRepo.GenerateSubscribeToken(ctx, userID, channel)
		if err != nil {
			uc.logger.Warn(ErrFailedToPublish.Error(),
				zap.Uint("threadID", id.ID),
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
