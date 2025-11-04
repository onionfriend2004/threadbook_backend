package dto

type UpdateProfileRequest struct {
	Nickname   string `json:"nickname,omitempty" validate:"omitempty,printascii,max=32"`
	AvatarLink string `json:"avatar_link,omitempty" validate:"omitempty,len=36"`
}

type UpdateProfileResponse struct {
	Username   string `json:"username,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarLink string `json:"avatar_link,omitempty"`
}
