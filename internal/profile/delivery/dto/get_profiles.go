package dto

type GetProfilesRequest struct {
	Usernames []string `json:"usernames" validate:"required,min=1,max=100,dive,required,alphanum,max=32"`
}

type GetProfilesResponseItem struct {
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	AvatarLink string `json:"avatar_link"`
}

type GetProfilesResponse struct {
	Profiles []GetProfilesResponseItem `json:"profiles"`
}
