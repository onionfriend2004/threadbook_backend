package dto

type InviteMemberInSpoolRequest struct {
	SpoolID         int      `json:"spool_id" binding:"required"`
	MemberUsernames []string `json:"member_usernames" binding:"required"`
}
