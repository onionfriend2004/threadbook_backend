package event

type Type string

const (
	// Message Events
	MessageCreated Type = "message.created"
	MessageUpdated Type = "message.updated"
	MessageDeleted Type = "message.deleted"

	// Thread Events
	ThreadCreated Type = "thread.created"
	ThreadUpdated Type = "thread.updated"
	ThreadDeleted Type = "thread.deleted"

	// Thread / Invite
	ThreadInvited Type = "thread.invited"

	// Spool Events
	SpoolUpdated Type = "spool.updated"
	SpoolDeleted Type = "spool.deleted"

	// Spool / Invite
	SpoolInvited Type = "spool.invited"
)

type Event struct {
	Type    Type `json:"type"`
	Payload any  `json:"payload"`
}

type ThreadSubTokenPayload struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

//
// ---- Message Events ----
//

type MessageCreatedPayload struct {
	MessageID uint     `json:"id"`
	ThreadID  uint     `json:"thread_id"`
	Content   string   `json:"content"`
	Username  string   `json:"username"`
	Payloads  []string `json:"payloads,omitempty"`
	CreatedAt int64    `json:"created_at"`
}

// для единообразия добавим и сюда Payloads — если при апдейте тоже есть файлы (например, после редактирования)
type MessageUpdatedPayload struct {
	MessageID uint   `json:"id"`
	ThreadID  uint   `json:"thread_id"`
	Content   string `json:"content"`
	// Payloads  []string `json:"payloads,omitempty"`
	UpdatedAt int64 `json:"updated_at"`
}

type MessageDeletedPayload struct {
	MessageID uint   `json:"id"`
	ThreadID  uint   `json:"thread_id"`
	DeletedBy string `json:"deleted_by,omitempty"`
}

// ---- Thread Events ----

type ThreadCreatedPayload struct {
	ThreadID       uint   `json:"id"`
	Title          string `json:"title"`
	CreatedAt      int64  `json:"created_at"`
	Channel        string `json:"channel"`
	Token          string `json:"token"`
	SubscribeToken string `json:"subscribe_token"`
}

type ThreadUpdatedPayload struct {
	ThreadID  uint   `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"`
}

type ThreadClosedPayload struct {
	ThreadID uint `json:"id"`
}

type ThreadInvitePayload struct {
	ThreadID uint   `json:"id"`
	Title    string `json:"title"`
	Channel  string `json:"channel"`
	Token    string `json:"token"`
}

//
// ---- Spool Events ----
//

type SpoolUpdatedPayload struct {
	SpoolID    uint   `json:"id"`
	BannerLink string `json:"banner_link,omitempty"`
	Name       string `json:"name"`
	UpdatedAt  int64  `json:"updated_at"`
}

type SpoolDeletedPayload struct {
	SpoolID   uint   `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
}

type SpoolInvitedPayload struct {
	SpoolID    uint   `json:"id"`
	BannerLink string `json:"banner_link,omitempty"`
	Name       string `json:"name"`
}
