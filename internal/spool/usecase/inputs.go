package usecase

// ---------- CreateSpool ----------
type CreateSpoolInput struct {
	OwnerID    uint
	Name       string
	BannerLink string
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
