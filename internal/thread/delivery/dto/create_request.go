package dto

type ThreadCreateRequest struct {
	Title      string `json:"title"`
	SpoolID    string `json:"spool_id"`
	TypeThread string `json:"type"`
}
