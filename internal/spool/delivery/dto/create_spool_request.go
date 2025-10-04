package dto

type CreateSpoolRequest struct {
	Name       string `json:"name" binding:"required"`
	BannerLink string `json:"banner_link,omitempty"`
}
