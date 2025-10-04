package dto

type GetSpoolInfoByIdRequest struct {
	SpoolID int `json:"spool_id" binding:"required"`
}
