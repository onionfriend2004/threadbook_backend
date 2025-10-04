package dto

type LeaveFromSpoolRequest struct {
	SpoolID int `json:"spool_id" binding:"required"`
}
