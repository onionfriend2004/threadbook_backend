package dto

type CheckEmailRequest struct {
	Email string `json:"email" validate:"required,max=32"`
}
