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

func (u *ThreadUsecase) GetVoiceToken(ctx context.Context, userID uint, username string, threadID int) (string, error) {
	if username == "" || threadID <= 0 {
		return "", ErrInvalidInput
	}

	thread, err := u.threadRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		return "", ErrThreadNotFound
	}

	hasRights, err := u.threadRepo.CheckRightsUserOnThreadRoom(ctx, thread.ID, userID)
	if !hasRights || err != nil {
		return "", ErrNoRightsOnJoinRoom
	}

	roomName := fmt.Sprintf("thread_%d", threadID) // Можно оптимизировать на 0,001% быстрее

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
		CanPublishSources: []string{"camera", "microphone", "screen"}, // Всё можно ж =)
	}
	// TODO: подумать над длительностью токена, захардкожу 15 минут
	token.SetVideoGrant(grant).SetIdentity(username).SetValidFor(15 * time.Minute)

	return token.ToJWT()
}
