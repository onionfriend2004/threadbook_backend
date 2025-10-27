package dto

type InviteRequest struct {
	ThreadID  int      `json:"thread_id"`
	Usernames []string `json:"invitee_usernames"`
}
