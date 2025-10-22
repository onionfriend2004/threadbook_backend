package dto

type ThreadCreateRequest struct {
	Title      string `json:"title"`
	SpoolID    uint   `json:"spool_id"`
	TypeThread string `json:"type"`
}
