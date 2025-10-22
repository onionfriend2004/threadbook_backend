package dto

type GetVoiceTokenRequest struct {
	ThreadID uint `json:"thread_id" binding:"required,min=1"`
}
