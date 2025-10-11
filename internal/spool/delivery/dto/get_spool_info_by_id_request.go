package dto

type GetSpoolInfoByIdRequest struct {
	SpoolID uint `json:"spool_id" binding:"required"`
}
