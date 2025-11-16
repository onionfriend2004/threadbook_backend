package dto

type ThreadUsersResponse struct {
	Username   string `json:"username"`
	IsGuest    bool   `json:"is_guest"`
	Nickname   string `json:"nickname"`
	AvatarLink string `json:"avatar_link"`
}
