package external

import (
	"context"
	"fmt"

	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LiveKitRepo struct {
	client          *lksdk.RoomServiceClient
	emptyRoomTTL    uint32
	maxParticipants uint32
}

func NewLiveKitRepo(client *lksdk.RoomServiceClient, emptyRoomTTL uint32, maxParticipants uint32) SFUInterface {
	return &LiveKitRepo{client: client, emptyRoomTTL: emptyRoomTTL, maxParticipants: maxParticipants}
}

// EnsureRoom обеспечивает пользователю комнату в треде
func (r *LiveKitRepo) EnsureRoom(ctx context.Context, roomName string) error {
	_, err := r.client.CreateRoom(
		ctx,
		&livekit.CreateRoomRequest{
			Name:            roomName,
			EmptyTimeout:    r.emptyRoomTTL,    // В секундах!!!
			MaxParticipants: r.maxParticipants, // Макс участнников в штуках <3
		},
	)
	fmt.Println(err)
	if err != nil {
		// Проверяем: ошибка "комната уже существует"?
		if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
			return nil // комната уже есть — это по правилам =)
		}
		// Любая другая ошибка — косячок
		return err
	}
	return nil
}
