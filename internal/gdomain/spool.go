package gdomain

import (
	"strings"
	"time"
)

type Spool struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"type:text;not null"`
	CreatorID  uint   `gorm:"column:creator_id;not null"`
	BannerLink string `gorm:"type:text" json:"banner_link,omitempty"`

	// связи
	Threads []Thread `gorm:"foreignKey:SpoolID;constraint:OnDelete:CASCADE;"`
	Members []User   `gorm:"many2many:user_spool;constraint:OnDelete:CASCADE;"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// UserSpool — join таблица пользователь <-> спул
type UserSpool struct {
	UserID    uint `gorm:"primaryKey"`
	SpoolID   uint `gorm:"primaryKey"`
	IsDeleted bool `gorm:"default:false"`
}

// NormalizeName приводит название к нормализованному виду
func NormalizeName(name string) string {
	return strings.TrimSpace(name)
}

// NewSpool конструктор для Spool
func NewSpool(name, bannerLink string, creatorID uint) (*Spool, error) {
	normName := NormalizeName(name)
	if normName == "" {
		return nil, ErrEmptyName
	}

	return &Spool{
		Name:       normName,
		BannerLink: strings.TrimSpace(bannerLink),
		CreatorID:  creatorID,
	}, nil
}

type SpoolWithCreator struct {
	ID         uint
	Name       string
	BannerLink string
	IsCreator  bool
}
