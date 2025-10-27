package external

import (
	"context"
	"errors"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
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
	InviteToThread(ctx context.Context, inviterID int, inviteeUsernames []string, threadID int) error
	Update(ctx context.Context, input domain.UpdateThreadInput) (*domain.Thread, error)
	GetThreadByID(ctx context.Context, threadID int) (*domain.Thread, error)

	CheckRightsUserOnThreadRoom(ctx context.Context, threadID int, userID uint) (bool, error)
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
			Table("user_spools").
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
				Table("user_spools").
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
					})
				}

				if err := tx.Table("thread_users").Create(&records).Error; err != nil {
					return err
				}
			}
		} else {
			// Иначе добавляем только создателя
			if err := tx.Table("thread_users").Create(map[string]interface{}{
				"user_id":   creatorID,
				"thread_id": thread.ID,
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
		Joins("JOIN thread_users ut ON ut.thread_id = t.id").
		Where("t.spool_id = ? AND ut.user_id = ?", spoolID, userID).
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

// TODO: OPTIMIZE INDEX SEARCH
// CREATE INDEX idx_thread_users_user_thread_member
// ON thread_users (user_id, thread_id)
// WHERE is_member = true;
func (r *ThreadRepository) CheckRightsUserOnThreadRoom(ctx context.Context, threadID int, userID uint) (bool, error) {
	var count int64
	err := r.Db.WithContext(ctx).
		Table("thread_users").
		Where("user_id = ? AND thread_id = ? AND is_member = ?", userID, threadID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ThreadRepository) InviteToThread(ctx context.Context, inviterID int, inviteeUsernames []string, threadID int) error {
	var thread gdomain.Thread
	if err := r.Db.First(&thread, threadID).Error; err != nil {
		return err
	}

	if thread.Type != "private" {
		return ErrUserNoAccess
	}

	if inviterID != thread.CreatorID {
		return ErrUserNoAccess
	}

	for _, username := range inviteeUsernames {
		var invitee gdomain.User
		if err := r.Db.Where("username = ?", username).First(&invitee).Error; err != nil {
			return err
		}

		// Проверяем, что пользователь уже в спуле потока
		var inSpool int64
		if err := r.Db.Table("user_spools").
			Where("user_id = ? AND spool_id = ?", invitee.ID, thread.SpoolID).
			Count(&inSpool).Error; err != nil {
			return err
		}
		if inSpool == 0 {
			return ErrUserNotInSpool
		}

		// Проверяем, что пользователь ещё не в потоке
		var exists int64
		if err := r.Db.Table("thread_users").
			Where("user_id = ? AND thread_id = ?", invitee.ID, thread.ID).
			Count(&exists).Error; err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		// Добавляем пользователя в поток
		if err := r.Db.Model(&thread).Association("Users").Append(&invitee); err != nil {
			return err
		}
	}

	return nil
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
