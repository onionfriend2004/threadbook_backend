package dto

type InviteMemberInSpoolRequest struct {
	SpoolID         uint     `json:"spool_id" binding:"required" validate:"required,gt=0"`
	MemberUsernames []string `json:"member_usernames" binding:"required" validate:"required,min=1,dive,min=3,max=16"`
}
