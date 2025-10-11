package dto

type GetSpoolMembersRequest struct {
	SpoolID uint `json:"spool_id" binding:"required"`
}
