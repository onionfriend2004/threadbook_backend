package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=32"`
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,password"`
}
