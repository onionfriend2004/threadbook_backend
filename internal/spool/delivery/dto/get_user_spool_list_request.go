package dto

type GetUserSpoolListRequest struct {
	UserID int `json:"user_id" binding:"required"`
}
