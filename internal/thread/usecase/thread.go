package usecase

import (
	"context"
	// "errors"

	"github.com/onionfriend2004/threadbook_backend/internal/thread/domain"
	repo "github.com/onionfriend2004/threadbook_backend/internal/thread/external"
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, title string, spool_id int, typeThread string) (*domain.Thread, error)
	GetBySpoolID(ctx context.Context, spool_id int) ([]*domain.Thread, error)
}

type ThreadUsecase struct {
	threadRepo repo.ThreadRepositoryInterface
	logger     *zap.Logger
}

func NewThreadUsecase(
	threadRepo repo.ThreadRepositoryInterface,
	logger *zap.Logger,
) ThreadUsecaseInterface {
	return &ThreadUsecase{
		threadRepo: threadRepo,
		logger:     logger,
	}
}

func (u *ThreadUsecase) CreateThread(ctx context.Context, title string, spool_id int, typeThread string) (*domain.Thread, error) {
	newThread, err := u.threadRepo.Create(ctx, title, spool_id, typeThread)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}

func (u *ThreadUsecase) GetBySpoolID(ctx context.Context, spool_id int) ([]*domain.Thread, error) {
	newThread, err := u.threadRepo.GetBySpoolID(ctx, spool_id)
	if err != nil {
		return nil, err
	}
	return newThread, nil
}
