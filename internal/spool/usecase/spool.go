package usecase

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/spool/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/external"

	"go.uber.org/zap"
)

type SpoolUsecaseInterface interface {
	CreateSpool(ctx context.Context, input CreateSpoolInput) (*domain.Spool, error)
	LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error
	GetUserSpoolList(ctx context.Context, userID int) ([]domain.Spool, error)
	InviteMemberInSpool(ctx context.Context, input InviteMemberInSpoolInput) error
	UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*domain.Spool, error)
	GetSpoolInfoById(ctx context.Context, spoolID int) (*domain.Spool, error)
	GetSpoolMembers(ctx context.Context, spoolID int) ([]domain.User, error)
}

type spoolUsecase struct {
	spoolRepo external.SpoolRepoInterface
	userRepo  external.UserRepoInterface
	logger    *zap.Logger
}

func NewSpoolUsecase(
	spoolRepo external.SpoolRepoInterface,
	userRepo external.UserRepoInterface,
	logger *zap.Logger,
) SpoolUsecaseInterface {
	return &spoolUsecase{
		spoolRepo: spoolRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

func (u *spoolUsecase) CreateSpool(ctx context.Context, input CreateSpoolInput) (*domain.Spool, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}

	newSpool, err := domain.NewSpool(input.Name, input.BannerLink)
	if err != nil {
		return nil, err
	}

	created, err := u.spoolRepo.CreateSpool(ctx, newSpool, input.OwnerID)
	if err != nil {
		u.logger.Error("failed to create spool", zap.Error(err))
		return nil, err
	}

	return created, nil
}

func (u *spoolUsecase) LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error {
	if input.UserID == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}
	return u.spoolRepo.RemoveUserFromSpool(ctx, input.UserID, input.SpoolID)
}

func (u *spoolUsecase) GetUserSpoolList(ctx context.Context, userID int) ([]domain.Spool, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	return u.spoolRepo.GetSpoolsByUser(ctx, userID)
}

func (u *spoolUsecase) InviteMemberInSpool(ctx context.Context, input InviteMemberInput) error {
	if input.UserID == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}
	return u.spoolRepo.AddUserToSpool(ctx, input.UserID, input.SpoolID)
}

func (u *spoolUsecase) UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*domain.Spool, error) {
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

func (u *spoolUsecase) GetSpoolMembers(ctx context.Context, spoolID int) ([]domain.User, error) {
	if spoolID == 0 {
		return nil, ErrInvalidInput
	}
	return u.spoolRepo.GetMembersBySpoolID(ctx, spoolID)
}

func (u *spoolUsecase) GetSpoolInfoById(ctx context.Context, spoolID int) (*domain.Spool, error) {
	if spoolID == 0 {
		return nil, ErrInvalidInput
	}

	spool, err := u.spoolRepo.GetSpoolByID(ctx, uint(spoolID))
	if err != nil {
		u.logger.Error("failed to get spool info", zap.Error(err), zap.Int("spool_id", spoolID))
		return nil, err
	}

	return spool, nil
}
