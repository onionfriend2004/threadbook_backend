package gdomain

import (
	"strings"
	"time"
)

type Spool struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string `gorm:"type:text;not null" json:"name"`
	BannerLink string `gorm:"type:text" json:"banner_link,omitempty"`

	// связи
	Threads []Thread `gorm:"foreignKey:SpoolID;constraint:OnDelete:CASCADE;" json:"threads,omitempty"`
	Members []User   `gorm:"many2many:user_spools;constraint:OnDelete:CASCADE;" json:"members,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
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
func NewSpool(name, bannerLink string) (*Spool, error) {
	normName := NormalizeName(name)
	if normName == "" {
		return nil, ErrEmptyName
	}

	return &Spool{
		Name:       normName,
		BannerLink: strings.TrimSpace(bannerLink),
	}, nil
}
