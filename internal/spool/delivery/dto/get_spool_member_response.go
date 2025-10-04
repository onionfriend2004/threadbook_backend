package dto

type GetSpoolMembersResponse struct {
	Members []MemberShortInfo `json:"members"`
}

type MemberShortInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar_link,omitempty"`
}
