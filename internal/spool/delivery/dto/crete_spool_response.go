package dto

type CreateSpoolResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
}
