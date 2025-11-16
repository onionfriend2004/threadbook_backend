package dto

type InviteRequest struct {
	Usernames []string `json:"invitee_usernames" validate:"required,min=1,max=100,dive,required,alphanum,max=32"`
}
