package usecase

// ---------- CreateSpool ----------
type CreateSpoolInput struct {
	OwnerID    uint   `json:"owner_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	BannerLink string `json:"banner_link,omitempty"`
}

// ---------- LeaveFromSpool ----------
type LeaveFromSpoolInput struct {
	UserID  int `json:"user_id" binding:"required"`
	SpoolID int `json:"spool_id" binding:"required"`
}

// ---------- GetUserSpoolList ----------
type GetUserSpoolListInput struct {
	UserID int `json:"user_id" binding:"required"`
}

// ---------- InviteMemberInSpool ----------
type InviteMemberInSpoolInput struct {
	SpoolID         int      `json:"spool_id" binding:"required"`
	MemberUsernames []string `json:"member_usernames" binding:"required"`
}

// ---------- UpdateSpool ----------
type UpdateSpoolInput struct {
	SpoolID    int    `json:"spool_id" binding:"required"`
	Name       string `json:"name,omitempty"`
	BannerLink string `json:"banner_link,omitempty"`
}

// ---------- GetSpoolInfoById ----------
type GetSpoolInfoByIdInput struct {
	SpoolID int `json:"spool_id" binding:"required"`
}

// ---------- GetSpoolMembers ----------
type GetSpoolMembersInput struct {
	SpoolID int `json:"spool_id" binding:"required"`
}
