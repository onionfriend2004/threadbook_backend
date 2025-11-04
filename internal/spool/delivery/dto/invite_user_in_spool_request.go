package dto

type InviteMemberInSpoolRequest struct {
	SpoolID         uint     `json:"spool_id" validate:"required,gte=1"`
	MemberUsernames []string `json:"member_usernames" validate:"required,min=1,max=100,dive,required,alphanum,max=32"`
}
