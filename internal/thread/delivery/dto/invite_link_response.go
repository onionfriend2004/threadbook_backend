package dto

import "github.com/onionfriend2004/threadbook_backend/internal/thread/external/domain"

type GetThreadInviteLinksResponse struct {
	InviteLinks []*domain.InviteLink `json:"invite_links"`
}
