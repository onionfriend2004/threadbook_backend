package dto

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"max=2000"`
}
