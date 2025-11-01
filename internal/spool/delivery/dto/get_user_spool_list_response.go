package dto

type GetUserSpoolListResponse struct {
	Spools []SpoolShortInfo `json:"spools"`
}

type SpoolShortInfo struct {
	SpoolID    uint   `json:"id"`
	Name       string `json:"name"`
	IsCreator  bool   `json:"is_creator"`
	BannerLink string `json:"banner_link,omitempty"`
}
