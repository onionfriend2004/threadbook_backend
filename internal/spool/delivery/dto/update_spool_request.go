package dto

type UpdateSpoolRequest struct {
	SpoolID    uint   `json:"spool_id" binding:"required"`
	Name       string `json:"name,omitempty"`
	BannerLink string `json:"banner_link,omitempty"`
}
