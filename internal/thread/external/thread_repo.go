package repo

import (
	"context"
	"database/sql"

	"github.com/onionfriend2004/threadbook_backend/internal/thread/domain"
	"go.uber.org/zap"
)

type ThreadRepositoryInterface interface {
	Create(ctx context.Context, title string, spool_id int, typeThread string) (*domain.Thread, error)
}

type ThreadRepository struct {
	Db     *sql.DB
	logger *zap.SugaredLogger
}

func NewThreadRepository(endPoint string, logger *zap.SugaredLogger) *ThreadRepository {
	threadRepo := &ThreadRepository{}
	db, _ := sql.Open("postgres", endPoint)
	// Обработать ошибку
	threadRepo.Db = db
	threadRepo.logger = logger
	return threadRepo
}

func (r *ThreadRepository) Create(ctx context.Context, title string, spoolID string, threadType string) (*domain.Thread, error) {
	const op = "ThreadRepository.Create"

	query := `
		INSERT INTO thread (title, spool_id, type)
		VALUES ($1, $2, $3)
        RETURNING id, spool_id, title, type, is_closed, created_at, updated_at
	`

	newThread := &domain.Thread{}

	r.logger.Debugw("Выполнение запроса на создание thread",
		"op", op,
		"title", title,
		"spool_id", spoolID,
		"type", threadType,
	)

	err := r.Db.QueryRow(
		query,
		title,
		spoolID,
		threadType,
	).Scan(
		&newThread.ID,
		&newThread.SpoolID,
		&newThread.Title,
		&newThread.Type,
		&newThread.IsClosed,
		&newThread.CreatedAt,
		&newThread.UpdatedAt,
	)
	if err != nil {
		r.logger.Errorw("Ошибка при создании нового thread",
			"op", op,
			"error", err,
			"title", title,
			"spool_id", spoolID,
		)
		return nil, err
	}

	r.logger.Infow("Новый thread успешно создан", "op", op, "title", title, "spool_id", spoolID)
	return newThread, nil
}
