package dto

type GetProfilesRequest struct {
	Usernames []string `json:"usernames"` // или Usernames []string `json:"usernames"`
}

type GetProfilesResponseItem struct {
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	AvatarLink string `json:"avatar_link"`
}

type GetProfilesResponse struct {
	Profiles []GetProfilesResponseItem `json:"profiles"`
}
