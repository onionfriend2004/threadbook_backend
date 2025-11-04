package dto

type AuthenticateResponse struct {
	Email      string `json:"email"`
	IsVerify   string `json:"is_verify"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	AvatarLink string `json:"avatar_link"`
	IsGuest    bool   `json:"is_guest"`
}
