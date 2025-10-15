package dto

type ThreadCreateRequest struct {
	Title      string `json:"title"`
	SpoolID    int    `json:"spool_id"`
	TypeThread string `json:"type"`
}
