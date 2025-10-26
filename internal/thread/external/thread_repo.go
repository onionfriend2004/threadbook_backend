package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type ThreadRepoInterface interface {
	Create(ctx context.Context, creatorID, spoolID uint, title, threadType string) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, userID, spoolID uint) ([]*gdomain.Thread, error)
	CloseThread(id, userID uint) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, inviterID, inviteeID, threadID uint) error
	Update(ctx context.Context, id, editorID uint, title *string, threadType *string) (*gdomain.Thread, error)
	GetThreadByID(ctx context.Context, threadID uint) (*gdomain.Thread, error)

	CheckRightsUserOnThreadRoom(ctx context.Context, threadID, userID uint) (bool, error)
	GetThreadMembers(ctx context.Context, threadID uint) ([]gdomain.ThreadUser, error)
	GetAccessibleThreadIDs(ctx context.Context, userID uint) ([]uint, error)
}
