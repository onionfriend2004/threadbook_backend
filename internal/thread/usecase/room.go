package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	liveKitAuth "github.com/livekit/protocol/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	spoolExternal "github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

var (
	CanPublish     = true
	CanSubscribe   = true
	CanPublishData = true
)

// Интерфейс usecase
type RoomUsecaseInterface interface {
	GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (GetVoiceTokenOutput, error)
}

// Структура usecase
type RoomUsecase struct {
	threadRepo  repo.ThreadRepoInterface
	liveKitRepo repo.SFUInterface
	spoolRepo   spoolExternal.SpoolRepoInterface
	liveKitURL  string
	apiKey      string
	apiSecret   string
	logger      *zap.Logger
}

// Конструктор
func NewRoomUsecase(
	threadRepo repo.ThreadRepoInterface,
	spoolRepo spoolExternal.SpoolRepoInterface,
	liveKitRepo repo.SFUInterface,
	liveKitURL, apiKey, apiSecret string,
	logger *zap.Logger,
) RoomUsecaseInterface {
	return &RoomUsecase{
		threadRepo:  threadRepo,
		spoolRepo:   spoolRepo,
		liveKitRepo: liveKitRepo,
		liveKitURL:  liveKitURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		logger:      logger,
	}
}

// Функция генерации TURN credentials по use-auth-secret
func generateTurnCredentials(identity, staticSecret string, ttl time.Duration) (username, credential string) {
	expire := time.Now().Add(ttl).Unix() // unix seconds
	username = fmt.Sprintf("%d:%s", expire, identity)

	h := hmac.New(sha1.New, []byte(staticSecret))
	h.Write([]byte(username))
	credential = base64.StdEncoding.EncodeToString(h.Sum(nil))

	return username, credential
}

// Структура вывода токена
type GetVoiceTokenOutput struct {
	Token    string   `json:"token"`
	TurnURLs []string `json:"turn_urls"`
	TurnUser string   `json:"turn_username"`
	TurnPass string   `json:"turn_credential"`
	TurnTTL  int64    `json:"turn_ttl_seconds"`
}

// Главная функция генерации токена
func (u *RoomUsecase) GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (GetVoiceTokenOutput, error) {
	var out GetVoiceTokenOutput

	// Проверяем тред
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return out, ErrThreadNotFound
	}
	if thread.IsClosed {
		return out, ErrThreadIsClosed
	}

	// Проверяем права доступа
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return out, ErrThreadNotFound
		}
		if !isMember {
			return out, ErrThreadNotFound
		}
	} else {
		userStatus, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return out, ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return out, ErrThreadNotFound
		}
	}

	roomName := fmt.Sprintf("thread_%d", input.ThreadID)

	// Создаём комнату, если её нет
	if err := u.liveKitRepo.EnsureRoom(ctx, roomName); err != nil {
		u.logger.Error("failed to ensure LiveKit room",
			zap.String("room_name", roomName),
			zap.Error(err),
		)
		return out, ErrFailedToEnsureRoom
	}

	// Генерируем JWT для LiveKit
	token := liveKitAuth.NewAccessToken(u.apiKey, u.apiSecret)
	grant := &liveKitAuth.VideoGrant{
		RoomJoin:          true,
		Room:              roomName,
		CanPublish:        &CanPublish,
		CanPublishData:    &CanPublishData,
		CanSubscribe:      &CanSubscribe,
		CanPublishSources: []string{"camera", "microphone", "screen"},
	}
	token.SetVideoGrant(grant).
		SetIdentity(input.Username).
		SetValidFor(15 * time.Minute)

	jwt, err := token.ToJWT()
	if err != nil {
		u.logger.Error("failed to generate LiveKit JWT", zap.Error(err))
		return out, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Генерируем TURN credentials с TTL = 15 минут
	staticAuthSecret := "mySuperSecret123" // можно вынести в env
	turnTTL := 15 * time.Minute
	turnUser, turnPass := generateTurnCredentials(input.Username, staticAuthSecret, turnTTL)

	out = GetVoiceTokenOutput{
		Token:    jwt,
		TurnURLs: []string{"turn:threadbook.ru:3478?transport=udp", "turns:threadbook.ru:5349?transport=tcp"},
		TurnUser: turnUser,
		TurnPass: turnPass,
		TurnTTL:  int64(turnTTL.Seconds()),
	}
	return out, nil
}
