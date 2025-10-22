package usecase

import (
	"context"
	"fmt"
	"time"

	liveKitAuth "github.com/livekit/protocol/auth"
)

var (
	CanPublish     = true
	CanSubscribe   = true
	CanPublishData = true
)

func (u *ThreadUsecase) GetVoiceToken(ctx context.Context, input GetVoiceTokenInput) (string, error) {
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
