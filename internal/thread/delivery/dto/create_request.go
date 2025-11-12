package dto

type ThreadCreateRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=32"`
	SpoolID     *uint  `json:"spool_id" validate:"omitempty,gt=0"`
	TypeThread  string `json:"type" validate:"required,oneof=public private"`
	AccessLevel uint   `json:"access_level" validate:"omitempty,gte=0,lte=3"`
}
