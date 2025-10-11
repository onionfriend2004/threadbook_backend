package dto

type UpdateSpoolResponse struct {
	SpoolID    uint   `json:"spool_id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
}
