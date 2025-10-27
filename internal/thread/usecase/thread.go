package usecase

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error)
	GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error)
	CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error)
	InviteToThread(ctx context.Context, input InviteToThreadInput) error
	UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error)
}

type ThreadUsecase struct {
	threadRepo repo.ThreadRepoInterface
	logger     *zap.Logger
}

func NewThreadUsecase(
	threadRepo repo.ThreadRepoInterface,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo: threadRepo,
		logger:     logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*gdomain.Thread, error) {
	if !(input.TypeThread == "private" || input.TypeThread == "public") {
		return nil, ErrWrognTypeThread
	}

	newThread, err := u.threadRepo.Create(ctx, input.OwnerID, input.SpoolID, input.Title, input.TypeThread)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) GetBySpoolID(ctx context.Context, input GetBySpoolIDInput) ([]*gdomain.Thread, error) {
	newThread, err := u.threadRepo.GetBySpoolID(ctx, input.UserID, input.SpoolID)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) CloseThread(ctx context.Context, input CloseThreadInput) (*gdomain.Thread, error) {
	return u.threadRepo.CloseThread(input.ThreadID, input.UserID)
}

func (u *ThreadUsecase) InviteToThread(ctx context.Context, input InviteToThreadInput) error {
	return u.threadRepo.InviteToThread(ctx, input.InviterID, input.InviteeUsernames, input.ThreadID)
}

func (u *ThreadUsecase) UpdateThread(ctx context.Context, input UpdateThreadInput) (*gdomain.Thread, error) {
	if input.ID == 0 {
		return nil, errors.New("thread id is required")
	}
	if input.EditorID == 0 {
		return nil, errors.New("editor id is required")
	}

	updatedThread, err := u.threadRepo.Update(ctx, input.ID, input.EditorID, input.Title, input.ThreadType)
	if err != nil {
		return nil, err
	}

	return updatedThread, nil
}
