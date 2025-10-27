package usecase

import (
	"context"
	"errors"
	"strconv"

	file "github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/profile/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/profile/external"
	"go.uber.org/zap"
)

type ProfileUsecaseInterface interface {
	UpdateProfile(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error)
	GetProfilesByUsernames(ctx context.Context, usernames []string) ([]dto.GetProfilesResponseItem, error)
}

type UpdateProfileOutput struct {
	UserID     int    `json:"user_id"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarLink string `json:"avatar_link,omitempty"`
}

type ProfileUsecase struct {
	profileRepo external.ProfileRepoInterface
	fileUC      file.FileUsecaseInterface
	logger      *zap.Logger
}

func NewProfileUsecase(profileRepo external.ProfileRepoInterface, fileUC file.FileUsecaseInterface, logger *zap.Logger) ProfileUsecaseInterface {
	return &ProfileUsecase{
		profileRepo: profileRepo,
		fileUC:      fileUC,
		logger:      logger,
	}
}

var (
	ErrEmptyProfileUpdate = errors.New("nothing to update")
)

func (uc *ProfileUsecase) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error) {
	// Если вообще ничего не передано — ошибка
	if input.Nickname == nil && input.Avatar == nil {
		return nil, ErrEmptyProfileUpdate
	}

	var avatarLink string
	var err error

	userIDstr := strconv.Itoa(input.UserID)

	// Если передан аватар, сохраняем через fileUC
	if input.Avatar != nil {
		avatarLink, err = uc.fileUC.SaveFile(ctx, file.SaveFile{
			File:        input.Avatar.File,
			Size:        input.Avatar.Size,
			Filename:    input.Avatar.Filename,
			ContentType: input.Avatar.ContentType,
			UserID:      userIDstr,
			FileType:    input.Avatar.Filetype,
		})
		if err != nil {
			uc.logger.Error("failed to save avatar", zap.Error(err))
			return nil, err
		}
	}

	// Обновляем профиль
	profile, err := uc.profileRepo.UpdateProfile(ctx, input.UserID, input.Nickname, avatarLink)
	if err != nil {
		uc.logger.Error("failed to update profile", zap.Error(err))
		return nil, err
	}

	// Формируем ответ
	output := &UpdateProfileOutput{
		UserID:     profile.ID,
		AvatarLink: profile.AvatarLink,
	}
	if input.Nickname != nil {
		output.Nickname = *input.Nickname
	}

	return output, nil
}

func (u *ProfileUsecase) GetProfilesByUsernames(ctx context.Context, usernames []string) ([]dto.GetProfilesResponseItem, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	users, err := u.profileRepo.GetProfilesByUsernames(ctx, usernames)
	if err != nil {
		u.logger.Error("failed to fetch users with profiles", zap.Error(err))
		return nil, err
	}

	profiles := make([]dto.GetProfilesResponseItem, 0, len(users))
	for _, user := range users {
		profiles = append(profiles, dto.GetProfilesResponseItem{
			Username:   user.Username,
			Nickname:   user.Profile.Nickname,
			AvatarLink: user.Profile.AvatarLink,
		})
	}

	return profiles, nil
}
