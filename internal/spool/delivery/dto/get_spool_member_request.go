package dto

type GetSpoolMembersRequest struct {
	SpoolID int `json:"spool_id" binding:"required"`
}
