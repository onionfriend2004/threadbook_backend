package external

import (
	"context"
	"errors"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/thread/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserNotInSpool   = errors.New("user not in spool")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUserNoAccess     = errors.New("user not owner")
)

type ThreadRepositoryInterface interface {
	Create(ctx context.Context, creatorID, spoolID int, title, threadType string) (*domain.Thread, error)
	GetBySpoolID(ctx context.Context, userID, spoolID int) ([]*domain.Thread, error)
	CloseThread(id int, userID int) (*domain.Thread, error)
	InviteToThread(ctx context.Context, inviterID, inviteeID, threadID int) error
	Update(ctx context.Context, input domain.UpdateThreadInput) (*domain.Thread, error)
	GetThreadByID(ctx context.Context, threadID int) (*domain.Thread, error)
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

func (r *ThreadRepository) Create(ctx context.Context, creatorID, spoolID int, title, threadType string) (*domain.Thread, error) {
	var thread domain.Thread

	err := r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.
			Table("user_spool").
			Where("user_id = ? AND spool_id = ?", creatorID, spoolID).
			Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			return ErrUserNotInSpool
		}

		// Создаём тред
		thread = domain.Thread{
			CreatorID: creatorID,
			SpoolID:   spoolID,
			Title:     title,
			Type:      threadType,
			IsClosed:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&thread).Error; err != nil {
			return err
		}

		if threadType == "public" {
			var userIDs []int
			if err := tx.
				Table("user_spool").
				Select("user_id").
				Where("spool_id = ?", spoolID).
				Pluck("user_id", &userIDs).Error; err != nil {
				return err
			}

			if len(userIDs) > 0 {
				records := make([]map[string]interface{}, 0, len(userIDs))
				for _, uid := range userIDs {
					records = append(records, map[string]interface{}{
						"user_id":   uid,
						"thread_id": thread.ID,
						"is_member": true,
					})
				}

				if err := tx.Table("user_thread").Create(&records).Error; err != nil {
					return err
				}
			}
		} else {
			// Иначе добавляем только создателя
			if err := tx.Table("user_thread").Create(map[string]interface{}{
				"user_id":   creatorID,
				"thread_id": thread.ID,
				"is_member": true,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &thread, nil
}

func (r *ThreadRepository) GetBySpoolID(ctx context.Context, userID, spoolID int) ([]*domain.Thread, error) {
	var threads []*domain.Thread
	const op = "ThreadRepository.GetBySpoolID"

	err := r.Db.
		Table("threads AS t").
		Joins("JOIN user_thread ut ON ut.thread_id = t.id").
		Where("t.spool_id = ? AND ut.user_id = ? AND ut.is_member = ?", spoolID, userID, true).
		Find(&threads).Error

	if err != nil {
		return nil, err
	}

	return threads, nil
}

func (r *ThreadRepository) CloseThread(id int, userID int) (*domain.Thread, error) {
	var thread domain.Thread
	if err := r.Db.First(&thread, id).Error; err != nil {
		return nil, err
	}
	if thread.CreatorID == userID {
		thread.IsClosed = true
		if err := r.Db.Save(&thread).Error; err != nil {
			return nil, err
		}
		return &thread, nil
	}
	return nil, ErrUserNoAccess
}

// DONT CHANGE THIS METHOD!!!
func (r *ThreadRepository) GetThreadByID(ctx context.Context, threadID int) (*domain.Thread, error) {
	var thread domain.Thread
	if err := r.Db.WithContext(ctx).First(&thread, threadID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	return &thread, nil
}

// DONT CHANGE THIS METHOD!!!

func (r *ThreadRepository) InviteToThread(ctx context.Context, inviterID, inviteeID, threadID int) error {
	// Проверяем, что тред приватный
	var thread struct {
		ID      int
		Type    string
		SpoolID int
	}
	if err := r.Db.
		Table("threads").
		Select("id, type, spool_id").
		Where("id = ?", threadID).
		Scan(&thread).Error; err != nil {
		return err
	}

	if thread.Type != "private" {
		return ErrUserNoAccess
	}

	var inThread int64
	if err := r.Db.
		Table("user_thread").
		Where("user_id = ? AND thread_id = ? AND is_member = true", inviterID, threadID).
		Count(&inThread).Error; err != nil {
		return err
	}
	if inThread == 0 {
		return ErrUserNoAccess
	}

	var inSpool int64
	if err := r.Db.
		Table("user_spool").
		Where("user_id = ? AND spool_id = ?", inviteeID, thread.SpoolID).
		Count(&inSpool).Error; err != nil {
		return err
	}
	if inSpool == 0 {
		return ErrUserNotInSpool
	}

	var exists int64
	if err := r.Db.
		Table("user_thread").
		Where("user_id = ? AND thread_id = ?", inviteeID, threadID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	return r.Db.Table("user_thread").Create(map[string]interface{}{
		"user_id":   inviteeID,
		"thread_id": threadID,
		"is_member": true,
	}).Error
}

func (r *ThreadRepository) Update(ctx context.Context, input domain.UpdateThreadInput) (*domain.Thread, error) {
	var thread domain.Thread

	err := r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем, существует ли тред
		if err := tx.First(&thread, "id = ?", input.ID).Error; err != nil {
			return err
		}

		if thread.CreatorID != input.EditorID {
			return ErrPermissionDenied
		}

		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		if input.Title != nil {
			updates["title"] = *input.Title
		}
		if input.Type != nil {
			updates["type"] = *input.Type
		}
		if input.IsClosed != nil {
			updates["is_closed"] = *input.IsClosed
		}

		if err := tx.Model(&thread).Updates(updates).Error; err != nil {
			return err
		}

		return tx.First(&thread, "id = ?", input.ID).Error
	})

	if err != nil {
		return nil, err
	}

	return &thread, nil
}
