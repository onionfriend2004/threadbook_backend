package usecase

import (
	"context"
	"fmt"
	"time"

	liveKitAuth "github.com/livekit/protocol/auth"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type RoomUsecaseInterface interface {
	GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error)
}
type RoomUsecase struct {
	threadRepo  repo.ThreadRepoInterface
	liveKitRepo repo.SFUInterface
	liveKitURL  string
	apiKey      string
	apiSecret   string
	logger      *zap.Logger
}

func NewRoomUsecase(
	threadRepo repo.ThreadRepoInterface,
	liveKitRepo repo.SFUInterface,
	liveKitURL, apiKey, apiSecret string,
	logger *zap.Logger,
) RoomUsecaseInterface {
	return &RoomUsecase{
		threadRepo:  threadRepo,
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
		return "", ErrFaildToEnsureRoom
	}

	token := liveKitAuth.NewAccessToken(u.apiKey, u.apiSecret)

	grant := &liveKitAuth.VideoGrant{
		RoomJoin:          true,
		Room:              roomName,
		CanPublish:        proto.Bool(true),
		CanPublishData:    proto.Bool(true),
		CanSubscribe:      proto.Bool(true),
		CanPublishSources: []string{"camera", "microphone", "audio", "screen"},
	}

	// TODO: подумать над длительностью токена, захардкожу 15 минут
	token.SetVideoGrant(grant).
		SetIdentity(input.Username).
		SetValidFor(15 * time.Minute)

	return token.ToJWT()
}
