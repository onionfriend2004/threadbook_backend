package dto

type CheckUsernameRequest struct {
	Username string `json:"username" validate:"required,alphanum,max=32"`
}
