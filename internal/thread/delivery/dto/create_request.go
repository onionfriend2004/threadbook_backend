package dto

type ThreadCreateRequest struct {
	Title      string `json:"title" validate:"required,min=3,max=32"`
	SpoolID    uint   `json:"spool_id" validate:"required,gt=0"`
	TypeThread string `json:"type" validate:"required,oneof=public private"`
}
