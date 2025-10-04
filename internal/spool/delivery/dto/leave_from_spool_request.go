package dto

type LeaveFromSpoolRequest struct {
	UserID  int `json:"user_id" binding:"required"`
	SpoolID int `json:"spool_id" binding:"required"`
}
