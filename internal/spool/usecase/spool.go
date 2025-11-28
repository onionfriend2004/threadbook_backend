package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	GetSpoolMembers(ctx context.Context, input GetSpoolMembersInput) ([]external.SpoolMember, error)
	AccessLevel(ctx context.Context, input AccessLevelInput) error

	GetSpoolInviteLinks(ctx context.Context, input GetSpoolInviteLinksInput) ([]*gdomain.InviteLink, error)
	DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error
	JoinToSpool(ctx context.Context, input JoinToSpoolInput) error
	CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*gdomain.InviteLink, error)
	RemoveAllGuestsFromSpool(ctx context.Context, input RemoveAllGuestsFromSpoolInput) error
}

type spoolUsecase struct {
	spoolRepo      external.SpoolRepoInterface
	inviteLinkRepo wsexternal.InviteLinkRepoInterface
	wsRepo         wsexternal.WebsocketRepoInterface
	fileUC         usecase.FileUsecaseInterface
	logger         *zap.Logger
}

func NewSpoolUsecase(
	spoolRepo external.SpoolRepoInterface,
	inviteLinkRepo wsexternal.InviteLinkRepoInterface,
	wsRepo wsexternal.WebsocketRepoInterface,
	fileUC usecase.FileUsecaseInterface,
	logger *zap.Logger,
) SpoolUsecaseInterface {
	return &spoolUsecase{
		spoolRepo:      spoolRepo,
		inviteLinkRepo: inviteLinkRepo,
		wsRepo:         wsRepo,
		fileUC:         fileUC,
		logger:         logger,
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
			return nil, ErrFailedToSaveBanner
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

	newSpool, err := gdomain.NewSpool(input.Name, bannerLink, input.OwnerID)
	if err != nil {
		return nil, ErrFailedToCreateSpool
	}

	var createdSpool *gdomain.Spool
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
		return nil, ErrFailedToCreateSpool
	}

	u.logger.Info("spool created successfully",
		zap.Uint("spool_id", createdSpool.ID),
		zap.String("spool_name", createdSpool.Name),
		zap.Bool("has_banner", bannerUploaded),
	)

	return createdSpool, nil
}

func (u *spoolUsecase) UpdateSpool(ctx context.Context, input UpdateSpoolInput) (*gdomain.Spool, error) {
	current, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil || current == nil {
		return nil, ErrSpoolNotFound
	}

	userStatus, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.UserID, input.SpoolID)
	if err != nil {
		return nil, err
	}
	if userStatus.AccessLevel < 3 {
		return nil, ErrNoAccessToSpool
	}

	var (
		newBannerLink string
		newUploaded   bool
	)

	if input.BannerInput != nil {
		fileInput := usecase.SaveFile{
			File:        input.BannerInput.File,
			Size:        input.BannerInput.Size,
			Filename:    input.BannerInput.Filename,
			ContentType: input.BannerInput.ContentType,
			UserID:      strconv.FormatUint(uint64(current.CreatorID), 10),
			FileType:    "spool_banner",
		}

		var saveErr error
		newBannerLink, saveErr = u.fileUC.SaveFile(ctx, fileInput)
		if saveErr != nil {
			return nil, ErrFailedToSaveBanner
		}
		newUploaded = true
	}

	defer func(link string, uploaded bool) {
		if uploaded && link != "" {
			if recovered := recover(); recovered != nil {
				u.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{Filename: link})
				panic(recovered)
			}
		}
	}(newBannerLink, newUploaded)

	var result *gdomain.Spool

	err = u.spoolRepo.WithTx(ctx, func(txCtx context.Context) error {
		bannerToSave := current.BannerLink
		if newUploaded {
			bannerToSave = newBannerLink
		}

		var txErr error
		result, txErr = u.spoolRepo.UpdateSpool(
			txCtx,
			input.SpoolID,
			input.Name,
			bannerToSave,
		)
		return txErr
	})

	if err != nil {
		u.logger.Error("failed to update spool", zap.Error(err))
		return nil, ErrFailedToUpdateSpool
	}

	if newUploaded && current.BannerLink != "" {
		if delErr := u.fileUC.DeleteFile(ctx, usecase.DeleteFileInput{
			Filename: current.BannerLink,
		}); delErr != nil {
			u.logger.Error("failed to delete old banner",
				zap.Error(delErr),
				zap.String("old_banner", current.BannerLink),
			)
		}
	}

	u.logger.Info("spool updated successfully",
		zap.Uint("spool_id", result.ID),
		zap.String("spool_name", result.Name),
		zap.Bool("banner_updated", newUploaded),
	)

	members, err := u.spoolRepo.GetMembersBySpoolID(ctx, result.ID)
	if err != nil {
		u.logger.Error("failed to get spool members",
			zap.Uint("spool_id", result.ID),
			zap.Error(err),
		)
	} else {
		payload := event.SpoolUpdatedPayload{
			SpoolID:    result.ID,
			BannerLink: result.BannerLink,
			Name:       result.Name,
			UpdatedAt:  time.Now().Unix(),
		}

		for _, member := range members {
			if err := u.wsRepo.PublishToUser(ctx, member.ID, event.Event{
				Type:    event.SpoolUpdated,
				Payload: payload,
			}); err != nil {
				u.logger.Warn("failed to publish SpoolUpdated event",
					zap.Uint("userID", member.ID),
					zap.Uint("spoolID", result.ID),
					zap.Error(err),
				)
			}
		}
	}

	return result, nil
}

// ---------- Leave ----------
func (u *spoolUsecase) LeaveFromSpool(ctx context.Context, input LeaveFromSpoolInput) error {
	if input.UserID == 0 || input.SpoolID == 0 {
		return ErrInvalidInput
	}

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to get spool info before leaving", zap.Error(err))
		return ErrInternal
	}

	if spool.CreatorID == input.UserID {
		u.logger.Warn("creator tried to leave their own spool",
			zap.Uint("creator_id", input.UserID),
			zap.Uint("spool_id", input.SpoolID),
		)
		return ErrForbidden
	}

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

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil || spool == nil {
		u.logger.Error("failed to get spool",
			zap.Uint("spool_id", input.SpoolID),
			zap.Error(err),
		)
		return ErrInternal
	}

	for _, username := range input.MemberUsernames {
		if username == "" {
			continue
		}

		// 1. Добавляем участника
		if err := u.spoolRepo.AddUserToSpoolByUsername(ctx, username, input.SpoolID); err != nil {
			u.logger.Error("failed to add user to spool",
				zap.String("username", username),
				zap.Error(err),
			)
			return ErrFailedToInvite
		}
	}

	// 2. Загружаем всех участников спула
	members, err := u.spoolRepo.GetMembersBySpoolID(ctx, input.SpoolID)
	if err != nil {
		u.logger.Error("failed to get spool members",
			zap.Uint("spool_id", input.SpoolID),
			zap.Error(err),
		)
		return ErrInternal
	}

	// 3. Готовим payload
	payload := event.SpoolInvitedPayload{
		SpoolID:    spool.ID,
		BannerLink: spool.BannerLink,
		Name:       spool.Name,
	}

	// 4. Широковещательная отправка ВСЕМ участникам
	for _, member := range members {
		if err := u.wsRepo.PublishToUser(ctx, member.ID, event.Event{
			Type:    event.SpoolInvited,
			Payload: payload,
		}); err != nil {
			u.logger.Warn("failed to publish SpoolInvited event",
				zap.Uint("userID", member.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// ---------- Get members ----------
func (u *spoolUsecase) GetSpoolMembers(ctx context.Context, input GetSpoolMembersInput) ([]external.SpoolMember, error) {
	if input.SpoolID == 0 || input.UserID == 0 {
		return nil, ErrInvalidInput
	}

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
	if err != nil {
		u.logger.Error("failed to get spool members from repository",
			zap.Uint("spool_id", input.SpoolID),
			zap.Error(err),
		)
		return nil, ErrFailedToGetMembers
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
		return nil, ErrFailedToGetSpool
	}

	return spool, nil
}

// ---------- InviteLinks ----------
func (u *spoolUsecase) CreateInviteLink(ctx context.Context, input CreateInviteLinkInput) (*gdomain.InviteLink, error) {

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if spool.CreatorID != input.UserID {
		return nil, ErrForbidden
	}
	session, err := u.inviteLinkRepo.CreateLink(ctx, "spool", input.SpoolID, input.UserID, input.ExpiresAt, input.MaxUses)
	if err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	return session, nil
}

func (u *spoolUsecase) JoinToSpool(ctx context.Context, input JoinToSpoolInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	link, err := u.inviteLinkRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrForbidden
	}

	return u.spoolRepo.AddUserToSpoolByUsername(ctx, input.Username, link.ResourceID)
}

func (u *spoolUsecase) DeleteInviteLink(ctx context.Context, input DeleteInviteLinkInput) error {
	if input.Link == "" {
		return ErrInvalidInput
	}

	link, err := u.inviteLinkRepo.GetInviteByID(ctx, input.Link)
	if err != nil {
		return ErrForbidden
	}

	spool, err := u.spoolRepo.GetSpoolByID(ctx, link.ResourceID)
	if err != nil || spool.CreatorID != input.UserID {
		return ErrForbidden
	}

	return u.inviteLinkRepo.DeleteLink(ctx, link.ID)
}

func (u *spoolUsecase) GetSpoolInviteLinks(ctx context.Context, input GetSpoolInviteLinksInput) ([]*gdomain.InviteLink, error) {

	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil || spool.CreatorID != input.UserID {
		return nil, ErrForbidden
	}

	return u.inviteLinkRepo.GetLinksByResource(ctx, "spool", input.SpoolID)
}

func (u *spoolUsecase) RemoveAllGuestsFromSpool(ctx context.Context, input RemoveAllGuestsFromSpoolInput) error {
	spool, err := u.spoolRepo.GetSpoolByID(ctx, input.SpoolID)
	if err != nil {
		return err
	}
	if spool.CreatorID != input.UserID {
		return ErrForbidden
	}
	return u.spoolRepo.RemoveAllGuestsFromSpool(ctx, input.SpoolID)
}

func (u *spoolUsecase) AccessLevel(ctx context.Context, input AccessLevelInput) error {
	editor, err := u.spoolRepo.GetUserSpoolStatus(ctx, input.EditorID, input.SpoolID)
	if err != nil {
		// TODO ПРОПИСАТЬ НОРМАЛЬНЫЕ ОШИБКИ
		return err
	}
	if editor == nil {
		return ErrInvalidInput
	}
	if editor.AccessLevel <= input.AccessLevel {
		// u.logger.Debug("editor access level <= input access level",
		// 	zap.Uint("editor.AccessLevel", editor.AccessLevel),
		// 	zap.Uint("input.AccessLevel", input.AccessLevel),
		// )
		// TODO ПРОПИСАТЬ НОРМАЛЬНЫЕ ОШИБКИ
		return ErrInvalidInput
	}
	user, err := u.spoolRepo.GetUserSpoolStatusByUsername(ctx, input.Username, input.SpoolID)
	if err != nil {
		return ErrInvalidInput
	}
	if user == nil {
		return ErrInvalidInput
	}
	if editor.AccessLevel <= user.AccessLevel {
		// u.logger.Debug("editor access level <= input user level",
		// 	zap.Uint("editor.AccessLevel", editor.AccessLevel),
		// 	zap.Uint("user.AccessLevel", user.AccessLevel),
		// )
		// TODO ПРОПИСАТЬ НОРМАЛЬНЫЕ ОШИБКИ
		return ErrInvalidInput
	}
	return u.spoolRepo.UpdateUserAccessLevel(ctx, user.UserID, input.SpoolID, input.AccessLevel)
}
