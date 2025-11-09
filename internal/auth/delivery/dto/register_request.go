package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=32"`
	Username string `json:"username" validate:"required,alphanum,max=32"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}
