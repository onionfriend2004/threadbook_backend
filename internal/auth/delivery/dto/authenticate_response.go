package dto

type AuthenticateResponse struct {
	Email    string `json:"email"`
	IsVerify string `json:"is_verify"`
	Username string `json:"username"`
}
