package dto

type CheckEmailRequest struct {
	Email string `json:"email" validate:"required,email,max=32"`
}
