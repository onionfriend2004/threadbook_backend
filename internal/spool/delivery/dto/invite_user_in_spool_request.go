package dto

type InviteMemberInSpoolRequest struct {
	SpoolID  int `json:"spool_id" binding:"required"`
	MemberID int `json:"member_id" binding:"required"`
}
