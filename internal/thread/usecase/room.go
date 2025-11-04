package usecase

import (
	"context"
	"fmt"
	"time"

	liveKitAuth "github.com/livekit/protocol/auth"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external/domain"
	"go.uber.org/zap"
)

var (
	CanPublish     = true
	CanSubscribe   = true
	CanPublishData = true
)

type RoomUsecaseInterface interface {
	GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error)
	CreateNoAuthSession(ctx context.Context, input CreateNoAuthSessionInput) (*domain.NoAuthSession, error)
	GetNoAuthVoiceToken(ctx context.Context, input GetNoAuthVoiceTokenInput) (string, error)
}
type RoomUsecase struct {
	threadRepo  repo.ThreadRepoInterface
	noAuthRepo  repo.NoAuthRepoInterface
	liveKitRepo repo.SFUInterface
	liveKitURL  string
	apiKey      string
	apiSecret   string
	logger      *zap.Logger
}

func NewRoomUsecase(
	threadRepo repo.ThreadRepoInterface,
	noAuthRepo repo.NoAuthRepoInterface,
	liveKitRepo repo.SFUInterface,
	liveKitURL, apiKey, apiSecret string,
	logger *zap.Logger,
) RoomUsecaseInterface {
	return &RoomUsecase{
		threadRepo:  threadRepo,
		noAuthRepo:  noAuthRepo,
		liveKitRepo: liveKitRepo,
		liveKitURL:  liveKitURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		logger:      logger,
	}
}

func (u *RoomUsecase) GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error) {
	if input.Username == "" || input.ThreadID <= 0 {
		return "", ErrInvalidInput
	}

	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return "", ErrThreadNotFound
	}

	hasRights, err := u.threadRepo.CheckRightsUserOnThreadRoom(ctx, thread.ID, input.UserID)
	if !hasRights || err != nil {
		return "", ErrNoRightsOnJoinRoom
	}

	roomName := fmt.Sprintf("thread_%d", input.ThreadID)

	if err := u.liveKitRepo.EnsureRoom(ctx, roomName); err != nil {
		u.logger.Error("failed to ensure LiveKit room",
			zap.String("room_name", roomName),
			zap.Error(err),
		)
		return "", ErrFailedToEnsureRoom
	}

	token := liveKitAuth.NewAccessToken(u.apiKey, u.apiSecret)

	grant := &liveKitAuth.VideoGrant{
		RoomJoin:          true,
		Room:              roomName,
		CanPublish:        &CanPublish,
		CanPublishData:    &CanPublishData,
		CanSubscribe:      &CanSubscribe,
		CanPublishSources: []string{"camera", "microphone", "screen"},
	}

	// TODO: подумать над длительностью токена, захардкожу 15 минут
	token.SetVideoGrant(grant).
		SetIdentity(input.Username).
		SetValidFor(15 * time.Minute)

	return token.ToJWT()
}

func (u *RoomUsecase) CreateNoAuthSession(ctx context.Context, input CreateNoAuthSessionInput) (*domain.NoAuthSession, error) {

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

	session, err := u.noAuthRepo.CreateNoAuthSession(ctx, input.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to create noauth session: %w", err)
	}

	return session, nil
}

func (u *RoomUsecase) GetNoAuthVoiceToken(ctx context.Context, input GetNoAuthVoiceTokenInput) (string, error) {
	if input.Nickname == "" || input.NoAuthSession == "" {
		return "", ErrInvalidInput
	}

	session, err := u.noAuthRepo.GetNoAuthSession(ctx, input.NoAuthSession)
	if err != nil {
		return "", ErrThreadNotFound
	}

	roomName := fmt.Sprintf("thread_%d", session.ThreadID)

	if err := u.liveKitRepo.EnsureRoom(ctx, roomName); err != nil {
		u.logger.Error("failed to ensure LiveKit room",
			zap.String("room_name", roomName),
			zap.Error(err),
		)
		return "", ErrFailedToEnsureRoom
	}

	token := liveKitAuth.NewAccessToken(u.apiKey, u.apiSecret)

	grant := &liveKitAuth.VideoGrant{
		RoomJoin:          true,
		Room:              roomName,
		CanPublish:        &CanPublish,
		CanPublishData:    &CanPublishData,
		CanSubscribe:      &CanSubscribe,
		CanPublishSources: []string{"camera", "microphone", "screen"},
	}

	// TODO: подумать над длительностью токена, захардкожу 15 минут
	token.SetVideoGrant(grant).
		SetIdentity(input.Nickname).
		SetValidFor(15 * time.Minute)

	return token.ToJWT()
}
