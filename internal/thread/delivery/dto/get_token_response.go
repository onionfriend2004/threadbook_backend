package dto

type GetVoiceTokenResponse struct {
	Token    string   `json:"token"`
	TurnURLs []string `json:"turn_urls"`
	TurnUser string   `json:"turn_username"`
	TurnPass string   `json:"turn_credential"`
	TurnTTL  int64    `json:"turn_ttl_seconds"`
}
