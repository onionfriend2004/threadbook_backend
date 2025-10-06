package infra

import (
	livekit "github.com/livekit/server-sdk-go/v2"
	"github.com/onionfriend2004/threadbook_backend/config"
)

func LiveKitConnect(cfg *config.Config) *livekit.RoomServiceClient {
	client := livekit.NewRoomServiceClient(
		cfg.LiveKit.URL,
		cfg.LiveKit.APIKey,
		cfg.LiveKit.APISecret,
	)
	return client
}
