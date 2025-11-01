package dto

type UpdateProfileRequest struct {
	Nickname   string `json:"nickname,omitempty"`
	AvatarLink string `json:"avatar_link,omitempty"`
}

type UpdateProfileResponse struct {
	Username   string `json:"username,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarLink string `json:"avatar_link,omitempty"`
}
