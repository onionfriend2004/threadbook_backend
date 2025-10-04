package usecase

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"go.uber.org/zap"
)

type ThreadUsecaseInterface interface {
	CreateThread(ctx context.Context, title string, spool_id int, typeThread string) (*domain.Thread, error)
}

type threadUsecase struct {
	threadRepo       external.ThreadRepoInterface
	logger         *zap.Logger
}

func NewThreadUsecase(
	threadRepo external.ThreadRepoInterface,
	logger *zap.Logger,
) AuthUsecaseInterface {
	return &authUsecase{
		threadRepo:       threadRepo,
		logger:         logger,
	}
}

func (*u ThreadUsecaseInterface) CreateThread(ctx context.Context, title string, spool_id int, typeThread string) {
	if newThread, err := u.threadRepo.Create(ctx context.Context, title string, spool_id int, typeThread string); err != nil {
		return nil, err
	}
	return newThread, nil
}
