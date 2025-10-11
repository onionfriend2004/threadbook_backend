package dto

type GetSpoolMembersResponse struct {
	Members []MemberShortInfo `json:"members"`
}

type MemberShortInfo struct {
	Username   string `json:"username"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarPath string `json:"avatar_link,omitempty"`
}
