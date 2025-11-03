package gdomain

type UserRegisteredEvent struct {
	Type     int    `json:"type"`
	Code     int    `json:"verify_code"`
	Email    string `json:"email_to"`
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
}
