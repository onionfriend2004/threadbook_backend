package dto

type InviteRequest struct {
	ThreadID  uint     `json:"thread_id" validate:"required,gt=0"`
	Usernames []string `json:"invitee_usernames" validate:"required,min=1,max=10,dive,min=3,max=16"`
}
