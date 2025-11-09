package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=32"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}
