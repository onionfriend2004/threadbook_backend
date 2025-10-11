package dto

type GetSpoolInfoByIdResponse struct {
	SpoolID    uint   `json:"spool_id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
