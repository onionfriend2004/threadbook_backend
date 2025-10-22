package usecase

import "github.com/onionfriend2004/threadbook_backend/internal/gdomain"

// ---------- CreateThread ----------
type CreateThreadInput struct {
	Title      string
	SpoolID    uint
	OwnerID    uint
	TypeThread string
}

// ---------- GetBySpoolID ----------
type GetBySpoolIDInput struct {
	UserID  uint
	SpoolID uint
}

// ---------- CloseThread ----------
type CloseThreadInput struct {
	ThreadID uint
	UserID   uint
}

// ---------- InviteToThread ----------
type InviteToThreadInput struct {
	InviterID uint
	InviteeID uint
	ThreadID  uint
}

// ---------- UpdateThread ----------
type UpdateThreadInput struct {
	ID         uint
	EditorID   uint
	Title      *string
	ThreadType *string
}

// ---------- GetVoiceToken ----------
type GetVoiceTokenInput struct {
	UserID   uint
	Username string
	ThreadID uint
}

// ---------- SendMessage ----------
type SendMessageInput struct {
	UserID   uint
	ThreadID uint
	Content  string
	Payloads []gdomain.MessagePayload
}

// ---------- GetMessages ----------
type GetMessagesInput struct {
	ThreadID uint
	Limit    int
	Offset   int
}

// ---------- GetSubscribeToken ----------
type GetSubscribeTokenInput struct {
	UserID   uint
	ThreadID uint
}
