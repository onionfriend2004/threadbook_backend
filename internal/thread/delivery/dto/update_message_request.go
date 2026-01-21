package dto

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}
