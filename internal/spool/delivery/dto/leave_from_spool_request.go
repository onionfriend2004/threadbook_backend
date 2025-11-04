package dto

type LeaveFromSpoolRequest struct {
	SpoolID uint `json:"spool_id" validate:"required,gte=1"`
}
