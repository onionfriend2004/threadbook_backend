package dto

type CheckUsernameRequest struct {
	Username string `json:"username" validate:"required,username"`
}
