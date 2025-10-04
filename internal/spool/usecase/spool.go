package usecase

import (
	"context"

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
	logger    *zap.Logger
}

func NewSpoolUsecase(
	spoolRepo external.SpoolRepoInterface,
	logger *zap.Logger,
) SpoolUsecaseInterface {
	return &spoolUsecase{
		spoolRepo: spoolRepo,
		logger:    logger,
	}
}

// ---------- Create ----------
func (u *spoolUsecase) CreateSpool(ctx context.Context, input CreateSpoolInput) (*gdomain.Spool, error) {
	if input.Name == "" || input.OwnerID == 0 {
		return nil, ErrInvalidInput
	}

	newSpool, err := gdomain.NewSpool(input.Name, input.BannerLink)
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
	if input.MemberID == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}
	return u.spoolRepo.AddUserToSpool(ctx, input.MemberID, input.SpoolID)
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

	spool, err := u.spoolRepo.GetSpoolByID(ctx, uint(input.SpoolID))
	if err != nil {
		u.logger.Error("failed to get spool info", zap.Error(err), zap.Int("spool_id", input.SpoolID))
		return nil, err
	}

	return spool, nil
}
