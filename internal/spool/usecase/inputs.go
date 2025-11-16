package usecase

import (
	"io"
	"time"
)

// ---------- CreateSpool ----------
type CreateSpoolInput struct {
	OwnerID     uint
	Name        string
	BannerInput *BannerInput
}

type BannerInput struct {
	File        io.Reader
	Size        int64
	Filename    string
	ContentType string
}

// ---------- GetUserSpoolList ----------
type GetUserSpoolListInput struct {
	UserID uint
}

// ---------- InviteMemberInSpool ----------
type InviteMemberInSpoolInput struct {
	UserID          uint
	Username        string
	SpoolID         uint
	MemberUsernames []string
}

// ---------- UpdateSpool ----------
type UpdateSpoolInput struct {
	UserID     uint
	SpoolID    uint
	Name       string
	BannerLink string
}

// ---------- LeaveFromSpool ----------
type LeaveFromSpoolInput struct {
	UserID  uint
	SpoolID uint
}

// ---------- GetSpoolInfoById ----------
type GetSpoolInfoByIdInput struct {
	UserID  uint
	SpoolID uint
}

// ---------- GetSpoolMembers ----------
type GetSpoolMembersInput struct {
	UserID  uint
	SpoolID uint
}

// ---------- InviteLinks ----------
type CreateInviteLinkInput struct {
	UserID    uint
	SpoolID   uint
	MaxUses   uint
	ExpiresAt time.Time
}

type JoinToSpoolInput struct {
	Username string
	Link     string
}

type DeleteInviteLinkInput struct {
	UserID uint
	Link   string
}

type GetSpoolInviteLinksInput struct {
	UserID  uint
	SpoolID uint
}

type RemoveAllGuestsFromSpoolInput struct {
	UserID  uint
	SpoolID uint
}

type AccessLevelInput struct {
	EditorID    uint
	SpoolID     uint
	Username    string
	AccessLevel uint
}
