package dto

type CreateSpoolResponse struct {
	SpoolID    uint   `json:"spool_id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
}
