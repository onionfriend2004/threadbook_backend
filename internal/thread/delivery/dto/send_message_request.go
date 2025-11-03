package dto

type SendMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}
