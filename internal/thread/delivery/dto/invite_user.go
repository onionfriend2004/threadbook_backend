package dto

type InviteRequest struct {
	ThreadID  uint     `json:"thread_id"`
	Usernames []string `json:"invitee_usernames"`
}
