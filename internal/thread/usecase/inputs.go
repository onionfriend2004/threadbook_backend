package usecase

import (
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
)

// ---------- CreateThread ----------
type CreateThreadInput struct {
	Title       string
	SpoolID     *uint
	OwnerID     uint
	ThreadType  gdomain.ThreadType
	AccessLevel *uint
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
	InviterID        uint
	InviteeUsernames []string
	ThreadID         uint
}

// ---------- UpdateThread ----------
type UpdateThreadInput struct {
	ThreadID    uint
	EditorID    uint
	Title       *string
	ThreadType  *string
	AccessLevel *uint
}

// ---------- GetVoiceToken ----------
type GetVoiceTokenInput struct {
	UserID   uint
	Username string
	ThreadID uint
}

// ---------- SendMessage ----------
type FilePayload struct {
	File        multipart.File
	Filename    string
	Size        int64
	ContentType string
}

type SendMessageInput struct {
	ThreadID uint
	UserID   uint
	Username string
	Content  string
	Payloads []*FilePayload
}

// ---------- GetMessages ----------
type GetMessagesInput struct {
	UserID   uint
	ThreadID uint
	CursorID uint // ID сообщения, относительно которого грузим
	Limit    int  // сколько сообщений загрузить
	Forward  bool // true — новые сообщения после курсора, false — старые перед курсором
}

type UpdateMessageInput struct {
	ThreadID  uint
	MessageID uint
	UserID    uint
	Content   string
}

type DeleteMessageInput struct {
	ThreadID  uint
	MessageID uint
	UserID    uint
}

// ---------- GetSubscribeToken ----------
type GetSubscribeTokenInput struct {
	UserID   uint
	ThreadID uint
}

// ---------- GetConnectAndSubscribeTokens ----------
type ConnectAndSubscribeTokens struct {
	ConnectToken  string
	ChannelTokens map[string]string
}

type CreateInviteLinkInput struct {
	UserID    uint
	ThreadID  uint
	MaxUses   uint
	ExpiresAt time.Time
}

type JoinToThreadInput struct {
	UserID      uint
	Username    string
	InviteToken string
}

type DeleteInviteLinkInput struct {
	UserID uint
	Link   string
}

type GetThreadInviteLinksInput struct {
	UserID   uint
	ThreadID uint
}

type GetThreadUsersInput struct {
	UserID   uint
	ThreadID uint
}
