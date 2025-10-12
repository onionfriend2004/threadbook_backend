package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onionfriend2004/threadbook_backend/internal/file/usecase"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/external"
	"go.uber.org/zap"
)

type SpoolUsecaseInterface interface {
	CreateSpool(ctx context.Context, input CreateSpoolInput) (*gdomain.Spool, error)
	LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error
	GetUserSpoolList(ctx context.Context, input GetUserSpoolListInput) ([]gdomain.Spool, error)
	InviteMemberInSpool(ctx context.Context, input InviteMemberInSpoolInput) error
	UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*gdomain.Spool, error)
	GetSpoolInfoById(ctx context.Context, input GetSpoolInfoByIdInput) (*gdomain.Spool, error)
	GetSpoolMembers(ctx context.Context, input GetSpoolMembersInput) ([]gdomain.User, error)
}

type spoolUsecase struct {
	spoolRepo external.SpoolRepoInterface
	fileUC    usecase.FileUsecaseInterface
	logger    *zap.Logger
}

func NewSpoolUsecase(
	spoolRepo external.SpoolRepoInterface,
	fileUC usecase.FileUsecaseInterface,
	logger *zap.Logger,
) SpoolUsecaseInterface {
	return &spoolUsecase{
		spoolRepo: spoolRepo,
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
			if uploaded {
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
	newSpool, err := gdomain.NewSpool(input.Name, bannerLink)
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
	return u.spoolRepo.RemoveUserFromSpool(ctx, input.UserID, input.SpoolID)
}

// ---------- List by user ----------
func (u *spoolUsecase) GetUserSpoolList(ctx context.Context, input GetUserSpoolListInput) ([]gdomain.Spool, error) {
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
	}
	return nil
}

// ---------- Update ----------
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
	if input.SpoolID == 0 {
		return nil, ErrInvalidInput
	}
	return u.spoolRepo.GetMembersBySpoolID(ctx, input.SpoolID)
}

// ---------- Get info ----------
func (u *spoolUsecase) GetSpoolInfoById(ctx context.Context, input GetSpoolInfoByIdInput) (*gdomain.Spool, error) {
	if input.SpoolID == 0 {
		return nil, ErrInvalidInput
	}

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to get spool info", zap.Error(err), zap.Uint("spool_id", input.SpoolID))
		return nil, err
	}

	return spool, nil
}
