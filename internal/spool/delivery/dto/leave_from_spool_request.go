package dto

type LeaveFromSpoolRequest struct {
	SpoolID uint `json:"spool_id" binding:"required"`
}
