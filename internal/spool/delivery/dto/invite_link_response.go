package dto

import "github.com/onionfriend2004/threadbook_backend/internal/gdomain"

type GetSpoolInviteLinksResponse struct {
	InviteLinks []*gdomain.InviteLink `json:"invite_links"`
}
