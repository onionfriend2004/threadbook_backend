package dto

type InviteRequest struct {
	ThreadID  uint `json:"thread_id"`
	InviteeID uint `json:"invitee_id"`
}
