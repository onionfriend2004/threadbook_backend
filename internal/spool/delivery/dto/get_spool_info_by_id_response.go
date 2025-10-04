package dto

type GetSpoolInfoByIdResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
