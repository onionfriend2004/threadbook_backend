package gdomain

type EmailEvent struct {
	Type  int    `json:"type"`
	Code  int    `json:"verify_code"`
	Email string `json:"email_to"`
}
