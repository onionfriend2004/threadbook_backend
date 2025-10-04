package dto

type ThreadCreateRequest struct {
	title    string `json:"title"`
	spoolID string `json:"spool_id"`
	typeThread string `json:"type"`
}
