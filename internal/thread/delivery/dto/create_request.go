package dto

import "github.com/onionfriend2004/threadbook_backend/internal/gdomain"

type ThreadCreateRequest struct {
	Title       string             `json:"title" validate:"required,min=3,max=32"`
	SpoolID     *uint              `json:"spool_id" validate:"omitempty,gt=0"`
	ThreadType  gdomain.ThreadType `json:"type" validate:"required,threadtype"`
	AccessLevel *uint              `json:"access_level" validate:"omitempty,gte=0,lte=3"`
}
