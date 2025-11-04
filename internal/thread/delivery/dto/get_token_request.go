package dto

type GetVoiceTokenRequest struct {
	ThreadID uint `json:"thread_id" validate:"required,gte=1"`
}
