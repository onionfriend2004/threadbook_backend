package dto

type GetSpoolMembersRequest struct {
	SpoolID uint `json:"spool_id" validate:"required,gte=1"` // gte == greather than or equal
}
