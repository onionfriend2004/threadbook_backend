package usecase

// ---------- CreateThread ----------
type CreateThreadInput struct {
	Title      string
	SpoolID    int
	OwnerID    int
	TypeThread string
}

// ---------- GetBySpoolID ----------
type GetBySpoolIDInput struct {
	UserID  int
	SpoolID int
}

// ---------- CloseThread ----------
type CloseThreadInput struct {
	ThreadID int
	UserID   int
}

// ---------- InviteToThread ----------
type InviteToThreadInput struct {
	InviterID int
	InviteeID int
	ThreadID  int
}

// ---------- UpdateThread ----------
type UpdateThreadInput struct {
	ID         int
	EditorID   int
	Title      *string
	ThreadType *string
}

// ---------- GetVoiceToken ----------
type GetVoiceTokenInput struct {
	UserID   uint
	Username string
	ThreadID int
}
