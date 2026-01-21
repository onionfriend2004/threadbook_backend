package dto

type SendMessageRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}
