package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/event"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	wsexternal "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type SpoolUsecaseInterface interface {
	CreateSpool(ctx context.Context, input CreateSpoolInput) (*gdomain.Spool, error)
	LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error
	GetUserSpoolList(ctx context.Context, input GetUserSpoolListInput) ([]gdomain.SpoolWithCreator, error)
	InviteMemberInSpool(ctx context.Context, input InviteMemberInSpoolInput) error
	UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*gdomain.Spool, error)
	GetSpoolInfoById(ctx context.Context, input GetSpoolInfoByIdInput) (*gdomain.Spool, error)
	GetSpoolMembers(ctx context.Context, input GetSpoolMembersInput) ([]gdomain.User, error)
}

type spoolUsecase struct {
	spoolRepo external.SpoolRepoInterface
	wsRepo    wsexternal.WebsocketRepoInterface
	fileUC    usecase.FileUsecaseInterface
	logger    *zap.Logger
}

func NewSpoolUsecase(
	spoolRepo external.SpoolRepoInterface,
	wsRepo wsexternal.WebsocketRepoInterface,
	fileUC usecase.FileUsecaseInterface,
	logger *zap.Logger,
) SpoolUsecaseInterface {
	return &spoolUsecase{
		spoolRepo: spoolRepo,
		wsRepo:    wsRepo,
		fileUC:    fileUC,
		logger:    logger,
	}
}

// ---------- Create ----------
func (u *spoolUsecase) CreateSpool(ctx context.Context, input CreateSpoolInput) (*gdomain.Spool, error) {
	var bannerLink string
	var bannerUploaded bool

	if input.BannerInput != nil {
		fileInput := usecase.SaveFile{
			File:        input.BannerInput.File,
			Size:        input.BannerInput.Size,
			Filename:    input.BannerInput.Filename,
			ContentType: input.BannerInput.ContentType,
			UserID:      strconv.FormatUint(uint64(input.OwnerID), 10),
			FileType:    "spool_banner",
		}

		var saveErr error
		bannerLink, saveErr = u.fileUC.SaveFile(ctx, fileInput)
		if saveErr != nil {
			return nil, fmt.Errorf("failed to save banner: %w", saveErr)
		}
		bannerUploaded = true

		defer func(bannerLink string, uploaded bool) {
			if !uploaded {
				if deleteErr := u.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{Filename: bannerLink}); deleteErr != nil {
					u.logger.Error("failed to cleanup banner after error",
						zap.Error(deleteErr),
						zap.String("banner_link", bannerLink),
					)
				}
			}
		}(bannerLink, bannerUploaded)
	}

	// Создаем доменную модель спула
	newSpool, err := gdomain.NewSpool(input.Name, bannerLink, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create spool domain: %w", err)
	}

	var createdSpool *gdomain.Spool
	// Оборачиваем сохранение в БД в транзакцию
	err = u.spoolRepo.WithTx(ctx, func(txCtx context.Context) error {
		var txErr error
		createdSpool, txErr = u.spoolRepo.CreateSpool(txCtx, newSpool, input.OwnerID)
		return txErr
	})
	if err != nil {
		u.logger.Error("failed to create spool in database",
			zap.Error(err),
			zap.String("spool_name", input.Name),
		)
		return nil, fmt.Errorf("failed to save spool to database: %w", err)
	}

	u.logger.Info("spool created successfully",
		zap.Uint("spool_id", createdSpool.ID),
		zap.String("spool_name", createdSpool.Name),
		zap.Bool("has_banner", bannerUploaded),
	)

	return createdSpool, nil
}

// ---------- Leave ----------
func (u *spoolUsecase) LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error {
	if input.UserID == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}

	// Проверяем, кто является создателем спула
	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to get spool info before leaving", zap.Error(err))
		return ErrInternal
	}

	// Создатель не может выйти из собственного спула
	if spool.CreatorID == input.UserID {
		u.logger.Warn("creator tried to leave their own spool",
			zap.Uint("creator_id", input.UserID),
			zap.Uint("spool_id", input.SpoolID),
		)
		return ErrForbidden
	}

	// Удаляем пользователя из спула
	if err := u.spoolRepo.RemoveUserFromSpool(ctx, input.UserID, input.SpoolID); err != nil {
		u.logger.Error("failed to remove user from spool", zap.Error(err))
		return ErrInternal
	}

	u.logger.Info("user left spool successfully",
		zap.Uint("user_id", input.UserID),
		zap.Uint("spool_id", input.SpoolID),
	)
	return nil
}

// ---------- List by user ----------
func (u *spoolUsecase) GetUserSpoolList(ctx context.Context, input GetUserSpoolListInput) ([]gdomain.SpoolWithCreator, error) {
	if input.UserID == 0 {
		return nil, ErrInvalidInput
	}
	return u.spoolRepo.GetSpoolsByUser(ctx, input.UserID)
}

// ---------- Invite ----------
func (u *spoolUsecase) InviteMemberInSpool(ctx context.Context, input InviteMemberInSpoolInput) error {
	if len(input.MemberUsernames) == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}

	for _, username := range input.MemberUsernames {
		if username == "" {
			continue
		}
		if err := u.spoolRepo.AddUserToSpoolByUsername(ctx, username, input.SpoolID); err != nil {
			u.logger.Error("failed to add user to spool", zap.String("username", username), zap.Error(err))
			return err
		}
		spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
		if err != nil {
			u.logger.Error("failed to get spool", zap.Uint("spool_id", input.SpoolID), zap.Error(err))
			return ErrInternal
		}

		payload := event.SpoolInvitedPayload{
			SpoolID:    spool.ID,
			BannerLink: spool.BannerLink,
			Name:       spool.Name,
		}

		if err := u.wsRepo.PublishToUser(ctx, input.UserID, event.Event{
			Type:    event.ThreadInvited,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish ThreadInvited event", zap.Uint("userID", input.UserID), zap.Error(err))
		}
	}
	return nil
}

// ---------- Update ----------
// надо переделать, если нужен будет
func (u *spoolUsecase) UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*gdomain.Spool, error) {
	if input.SpoolID == 0 {
		return nil, ErrInvalidInput
	}

	updated, err := u.spoolRepo.UpdateSpool(ctx, input.SpoolID, input.Name, input.BannerLink)
	if err != nil {
		u.logger.Error("failed to update spool", zap.Error(err))
		return nil, err
	}
	return updated, nil
}

// ---------- Get members ----------
func (u *spoolUsecase) GetSpoolMembers(ctx context.Context, input GetSpoolMembersInput) ([]gdomain.User, error) {
	if input.SpoolID == 0 || input.UserID == 0 {
		return nil, ErrInvalidInput
	}

	// Проверяем, что пользователь состоит в спуле
	inSpool, err := u.spoolRepo.IsUserInSpool(ctx, input.UserID, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to check user membership in spool", zap.Error(err))
		return nil, ErrInternal
	}
	if !inSpool {
		u.logger.Warn("user tried to access members without membership",
			zap.Uint("user_id", input.UserID),
			zap.Uint("spool_id", input.SpoolID),
		)

		return nil, ErrForbidden
	}

	members, err := u.spoolRepo.GetMembersBySpoolID(ctx, input.SpoolID)

	// Логируем результат
	if err != nil {
		u.logger.Error("failed to get spool members from repository",
			zap.Uint("spool_id", input.SpoolID),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	u.logger.Debug("successfully retrieved spool members",
		zap.Uint("spool_id", input.SpoolID),
		zap.Int("members_count", len(members)),
	)

	return members, nil
}

// ---------- Get info ----------
func (u *spoolUsecase) GetSpoolInfoById(ctx context.Context, input GetSpoolInfoByIdInput) (*gdomain.Spool, error) {
	if input.SpoolID == 0 || input.UserID == 0 {
		return nil, ErrInvalidInput
	}

	inSpool, err := u.spoolRepo.IsUserInSpool(ctx, input.UserID, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to check user membership in spool", zap.Error(err))
		return nil, ErrInternal
	}
	if !inSpool {
		u.logger.Debug("user tried to get spool info without membership",
			zap.Uint("user_id", input.UserID),
			zap.Uint("spool_id", input.SpoolID),
		)
		return nil, ErrForbidden
	}

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to get spool info", zap.Error(err))
		return nil, err
	}

	return spool, nil
}
