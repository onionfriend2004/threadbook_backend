package dto

type UpdateThreadRequest struct {
	ID    uint    `json:"id" validate:"required,gte=1"`
	Title *string `json:"title" validate:"omitempty,min=3,max=32"`
	Type  *string `json:"type" validate:"omitempty,oneof=public private"`
}
