package usecase

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/thread/domain"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

var (
	ErrWrognTypeThread = errors.New("wrong type of thread")
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, title string, spool_id int, owner_id int, typeThread string) (*domain.Thread, error)
	GetBySpoolID(ctx context.Context, userID, spoolID int) ([]*domain.Thread, error)
	CloseThread(ctx context.Context, id int, userID int) (*domain.Thread, error)
	InviteToThread(ctx context.Context, inviterID, inviteeID, threadID int) error
	UpdateThread(ctx context.Context, input domain.UpdateThreadInput) (*domain.Thread, error)
	GetVoiceToken(ctx context.Context, userID uint, username string, threadID int) (string, error)
}

type ThreadUsecase struct {
	threadRepo  repo.ThreadRepositoryInterface
	liveKitRepo repo.SFUInterface
	liveKitURL  string
	apiKey      string
	apiSecret   string
	logger      *zap.Logger
}

func NewThreadUsecase(
	threadRepo repo.ThreadRepositoryInterface,
	liveKitRepo repo.SFUInterface,
	liveKitURL, apiKey, apiSecret string,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo:  threadRepo,
		liveKitRepo: liveKitRepo,
		liveKitURL:  liveKitURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		logger:      logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, title string, spoolID int, ownerID int, typeThread string) (*domain.Thread, error) {
	if !(typeThread == "private" || typeThread == "public") {
		return nil, ErrWrognTypeThread
	}
	newThread, err := u.threadRepo.Create(ctx, ownerID, spoolID, title, typeThread)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) GetBySpoolID(ctx context.Context, userID, spoolID int) ([]*domain.Thread, error) {
	newThread, err := u.threadRepo.GetBySpoolID(ctx, userID, spoolID)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) CloseThread(ctx context.Context, id int, userID int) (*domain.Thread, error) {
	return u.threadRepo.CloseThread(id, userID)
}

func (u *ThreadUsecase) InviteToThread(ctx context.Context, inviterID, inviteeID, threadID int) error {
	err := u.threadRepo.InviteToThread(ctx, inviterID, inviteeID, threadID)
	if err != nil {
		return err
	}
	return nil
}

func (u *ThreadUsecase) UpdateThread(ctx context.Context, input domain.UpdateThreadInput) (*domain.Thread, error) {

	if input.ID == 0 {
		return nil, errors.New("thread id is required")
	}

	if input.EditorID == 0 {
		return nil, errors.New("editor id is required")
	}

	updatedThread, err := u.threadRepo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	return updatedThread, nil
}
