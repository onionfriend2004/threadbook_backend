package dto

type InviteRequest struct {
	ThreadID  uint     `json:"thread_id" validate:"required,gte=1"`
	Usernames []string `json:"invitee_usernames" validate:"required,min=1,max=100,dive,required,alphanum,max=32"`
}
