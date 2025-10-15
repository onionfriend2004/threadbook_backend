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

func (u *ThreadUsecase) GetVoiceToken(ctx context.Context, username string, threadID int) (string, error) {
	if username == "" || threadID <= 0 {
		fmt.Println("ERROR IS HERE!")
		return "", ErrInvalidInput
	}

	_, err := u.threadRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		fmt.Println(err)
		return "", ErrThreadNotFound
	}

	roomName := fmt.Sprintf("thread_%d", threadID) // Можно оптимизировать на 0,001% быстрее

	if err := u.liveKitRepo.EnsureRoom(ctx, roomName); err != nil {
		fmt.Println(err)
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
