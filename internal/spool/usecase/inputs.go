package usecase

import "io"

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

// ---------- LeaveFromSpool ----------
type LeaveFromSpoolInput struct {
	UserID  uint
	SpoolID uint
}

// ---------- GetUserSpoolList ----------
type GetUserSpoolListInput struct {
	UserID uint
}

// ---------- InviteMemberInSpool ----------
type InviteMemberInSpoolInput struct {
	SpoolID         uint
	MemberUsernames []string
}

// ---------- UpdateSpool ----------
type UpdateSpoolInput struct {
	SpoolID    uint
	Name       string
	BannerLink string
}

// ---------- GetSpoolInfoById ----------
type GetSpoolInfoByIdInput struct {
	SpoolID uint
}

// ---------- GetSpoolMembers ----------
type GetSpoolMembersInput struct {
	SpoolID uint
}
