package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	liveKitAuth "github.com/livekit/protocol/auth"
)

var (
	CanPublish     = true
	CanSubscribe   = true
	CanPublishData = true
)

func (u *ThreadUsecase) GetVoiceToken(ctx context.Context, userID int, threadID int) (string, error) {
	if userID <= 0 || threadID <= 0 {
		return "", ErrInvalidInput
	}

	_, err := u.threadRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		return "", ErrThreadNotFound
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
	token.SetVideoGrant(grant).SetIdentity(strconv.Itoa(userID)).SetValidFor(15 * time.Minute)

	return token.ToJWT()
}
