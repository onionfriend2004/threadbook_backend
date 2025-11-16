package usecase

import (
	"context"
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

type RoomUsecaseInterface interface {
	GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error)
	// CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*domain.InviteLink, error)
	// GetNoAuthVoiceToken(ctx context.Context, input GetInviteLinkInput) (string, error)
}
type RoomUsecase struct {
	threadRepo  repo.ThreadRepoInterface
	liveKitRepo repo.SFUInterface
	spoolRepo   spoolExternal.SpoolRepoInterface
	liveKitURL  string
	apiKey      string
	apiSecret   string
	logger      *zap.Logger
}

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

func (u *RoomUsecase) GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error) {
	thread, err := u.threadRepo.GetThreadByID(ctx, input.ThreadID)
	if err != nil {
		return "", ErrThreadNotFound
	}
	if thread.IsClosed {
		return "nil", ErrThreadIsClosed
	}
	if thread.Type == gdomain.ThreadTypePrivate {
		isMember, err := u.threadRepo.IsUserThreadMember(ctx, input.UserID, thread.ID)
		if err != nil {
			return "", ErrThreadNotFound
		}
		if !isMember {
			return "", ErrThreadNotFound
		}
	} else {
		userStatus, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, *thread.SpoolID)
		if err != nil {
			return "", ErrThreadNotFound
		}
		if userStatus.AccessLevel < thread.AccessLevel || userStatus.IsDeleted {
			return "", ErrThreadNotFound
		}
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
