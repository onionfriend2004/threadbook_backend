package repo

import (
	"context"

	"github.com/onionfriend2004/threadbook_backend/internal/thread/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ThreadRepositoryInterface interface {
	Create(ctx context.Context, title string, spool_id int, typeThread string) (*domain.Thread, error)
	GetBySpoolID(ctx context.Context, spoolID int) ([]*domain.Thread, error)
	CloseThread(id int) (*domain.Thread, error)
}

type ThreadRepository struct {
	Db     *gorm.DB
	logger *zap.Logger
}

func NewThreadRepository(db *gorm.DB, logger *zap.Logger) *ThreadRepository {
	return &ThreadRepository{
		Db:     db,
		logger: logger,
	}
}

func (r *ThreadRepository) Create(ctx context.Context, title string, spoolID int, threadType string) (*domain.Thread, error) {
	const op = "ThreadRepository.Create"

	newThread := &domain.Thread{
		Title:   title,
		SpoolID: spoolID,
		Type:    threadType,
	}

	// Используем GORM для создания записи
	if err := r.Db.Create(newThread).Error; err != nil {
		return nil, err
	}

	return newThread, nil
}

func (r *ThreadRepository) GetBySpoolID(ctx context.Context, spoolID int) ([]*domain.Thread, error) {
	var threads []*domain.Thread
	const op = "ThreadRepository.GetBySpoolID"
	if err := r.Db.Where("spool_id = ?", spoolID).Find(&threads).Error; err != nil {
		return nil, err
	}
	return threads, nil
}

const op = "ThreadRepository.CloseThread"

func (r *ThreadRepository) CloseThread(id int) (*domain.Thread, error) {
	var thread domain.Thread
	if err := r.Db.First(&thread, id).Error; err != nil {
		return nil, err
	}
	thread.IsClosed = true
	if err := r.Db.Save(&thread).Error; err != nil {
		return nil, err
	}
	return &thread, nil
}
