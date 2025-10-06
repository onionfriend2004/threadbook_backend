package dto

type GetVoiceTokenRequest struct {
	ThreadID int `json:"thread_id" binding:"required,min=1"`
}
