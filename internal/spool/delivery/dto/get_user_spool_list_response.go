package dto

type GetUserSpoolListResponse struct {
	Spools []SpoolShortInfo `json:"spools"`
}

type SpoolShortInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	BannerLink string `json:"banner_link,omitempty"`
}
