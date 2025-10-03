package dto

type AuthenticateResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}
