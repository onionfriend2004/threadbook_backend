package dto

type UpdateSpoolRequest struct {
	ID         int    `json:"id" binding:"required"`
	Name       string `json:"name,omitempty"`
	BannerLink string `json:"banner_link,omitempty"`
}
