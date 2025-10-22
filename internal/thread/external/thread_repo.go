package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type ThreadRepositoryInterface interface {
	Create(ctx context.Context, creatorID, spoolID int, title, threadType string) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, userID, spoolID int) ([]*gdomain.Thread, error)
	CloseThread(id int, userID int) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, inviterID, inviteeID, threadID int) error
	Update(ctx context.Context, id int, editorID int, title *string, threadType *string) (*gdomain.Thread, error)
	GetThreadByID(ctx context.Context, threadID int) (*gdomain.Thread, error)

	CheckRightsUserOnThreadRoom(ctx context.Context, threadID int, userID uint) (bool, error)
}
