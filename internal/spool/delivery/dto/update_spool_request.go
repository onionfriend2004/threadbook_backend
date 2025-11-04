package dto

type UpdateSpoolRequest struct {
	SpoolID    uint   `json:"spool_id" validate:"required,gte=1"`
	Name       string `json:"name,omitempty" validate:"omitempty,printascii,min=1,max=32"`
	BannerLink string `json:"banner_link,omitempty" validate:"omitempty,len=36"`
}
