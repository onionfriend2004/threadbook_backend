package dto

type InviteRequest struct {
	ThreadID  int `json:"thread_id"`
	InviteeID int `json:"invitee_id"`
}
