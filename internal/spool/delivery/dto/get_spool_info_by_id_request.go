package dto

type GetSpoolInfoByIdRequest struct {
	SpoolID uint `json:"spool_id" validate:"required,gte=1"`
}
