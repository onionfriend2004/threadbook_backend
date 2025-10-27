package external

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"gorm.io/gorm"
)

type ProfileRepoInterface interface {
	UpdateProfile(ctx context.Context, userID int, nickname *string, avatarLink string) (*gdomain.Profile, error)
	GetProfileByUserID(ctx context.Context, userID int) (*gdomain.Profile, error)
	GetProfilesByUsernames(ctx context.Context, usernames []string) ([]gdomain.User, error)
}

type ProfileRepo struct {
	db *gorm.DB
}

func NewProfileRepo(db *gorm.DB) ProfileRepoInterface {
	return &ProfileRepo{db: db}
}

func (r *ProfileRepo) UpdateProfile(ctx context.Context, userID int, nickname *string, avatarLink string) (*gdomain.Profile, error) {
	if userID == 0 {
		return nil, errors.New("invalid user id")
	}

	var profile gdomain.Profile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = gdomain.Profile{
			UserID:     userID,
			Nickname:   "",
			AvatarLink: "",
		}
		if nickname != nil {
			profile.Nickname = *nickname
		}
		if avatarLink != "" {
			profile.AvatarLink = avatarLink
		}
		if err := r.db.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, err
		}
		return &profile, nil
	} else if err != nil {
		return nil, err
	}

	updateData := make(map[string]interface{})
	if nickname != nil && profile.Nickname != *nickname {
		updateData["nickname"] = *nickname
		profile.Nickname = *nickname
	}
	if avatarLink != "" && profile.AvatarLink != avatarLink {
		updateData["avatar_link"] = avatarLink
		profile.AvatarLink = avatarLink
	}

	if len(updateData) > 0 {
		if err := r.db.WithContext(ctx).
			Model(&gdomain.Profile{}).
			Where("user_id = ?", userID).
			Updates(updateData).Error; err != nil {
			return nil, err
		}
	}

	return &profile, nil
}

func (r *ProfileRepo) GetProfileByUserID(ctx context.Context, userID int) (*gdomain.Profile, error) {
	var profile gdomain.Profile
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepo) GetProfilesByUsernames(ctx context.Context, usernames []string) ([]gdomain.User, error) {
	var users []gdomain.User

	// Загружаем пользователей и их профили (если есть)
	if err := r.db.WithContext(ctx).
		Preload("Profile").
		Where("username IN ?", usernames).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
