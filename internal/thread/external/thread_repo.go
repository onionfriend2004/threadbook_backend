package external

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

type ThreadRepoInterface interface {
	Create(ctx context.Context, creatorID uint, spoolID *uint, title string, threadType gdomain.ThreadType, accessLevel *uint) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, userID, spoolID uint) ([]*gdomain.Thread, error)
	CloseThread(ctx context.Context, id uint) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, inviterID uint, inviteeUsernames []string, threadID uint) error
	Update(ctx context.Context, id uint, title *string, threadType *string, accessLevel *uint) (*gdomain.Thread, error)
	GetThreadByID(ctx context.Context, threadID uint) (*gdomain.Thread, error)
	IsUserThreadMember(ctx context.Context, userID uint, threadID uint) (bool, error)
	GetThreadUsers(ctx context.Context, threadID uint) ([]gdomain.User, error)

	// CheckRightsUserOnThreadRoom(ctx context.Context, threadID, userID uint) (bool, error)
	GetUserThreadAccessLevelsToUpdate(ctx context.Context, userID uint, threadID uint) (userAccessLevel uint, threadAccessLevel uint, err error)
	GetUsersWithAccess(ctx context.Context, spoolID uint, minAccessLevel uint) ([]gdomain.User, error)
	GetAccessibleThreadIDs(ctx context.Context, userID uint) ([]uint, error)
	// GetAccessibleThreadIDsBySpool(ctx context.Context, userID, spoolID uint) ([]uint, error)

	IsThreadOwner(ctx context.Context, userID, threadID uint) (bool, error)
	AddUserToThread(ctx context.Context, userID, threadID uint) error
	// GetUserSpoolStatus(ctx context.Context, userID uint, spoolID uint) (*gdomain.UserSpool, error)
}
