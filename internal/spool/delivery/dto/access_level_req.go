package dto

type AccessLevelRequest struct {
	AccessLevel uint `json:"access_level" validate:"required,gte=0"`
}
