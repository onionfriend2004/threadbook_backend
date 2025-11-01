package external

import (
	"context"
	"errors"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ThreadRepo struct {
	Db     *gorm.DB
	logger *zap.Logger
}

func NewThreadRepo(db *gorm.DB, logger *zap.Logger) ThreadRepoInterface {
	return &ThreadRepo{
		Db:     db,
		logger: logger,
	}
}

func (r *ThreadRepo) Create(ctx context.Context, creatorID, spoolID uint, title, threadType string) (*gdomain.Thread, error) {
	var thread gdomain.Thread

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
		thread = gdomain.Thread{
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

func (r *ThreadRepo) GetBySpoolID(ctx context.Context, userID, spoolID uint) ([]*gdomain.Thread, error) {
	var threads []*gdomain.Thread
	const op = "ThreadRepo.GetBySpoolID"

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

func (r *ThreadRepo) CloseThread(id uint, userID uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread
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
func (r *ThreadRepo) GetThreadByID(ctx context.Context, threadID uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread
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
func (r *ThreadRepo) CheckRightsUserOnThreadRoom(ctx context.Context, threadID uint, userID uint) (bool, error) {
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

func (r *ThreadRepo) InviteToThread(ctx context.Context, inviterID, inviteeID, threadID uint) error {
	var thread struct {
		ID      uint
		Type    string
		SpoolID uint
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
		Table("thread_users").
		Where("user_id = ? AND thread_id = ?", inviterID, threadID).
		Count(&inThread).Error; err != nil {
		return err
	}
	if inThread == 0 {
		return ErrUserNoAccess
	}

	var inSpool int64
	if err := r.Db.
		Table("user_spools").
		Where("user_id = ? AND spool_id = ?", inviteeID, thread.SpoolID).
		Count(&inSpool).Error; err != nil {
		return err
	}
	if inSpool == 0 {
		return ErrUserNotInSpool
	}

	var exists int64
	if err := r.Db.
		Table("thread_users").
		Where("user_id = ? AND thread_id = ?", inviteeID, threadID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	return r.Db.Table("thread_users").Create(map[string]interface{}{
		"user_id":   inviteeID,
		"thread_id": threadID,
	}).Error
}

func (r *ThreadRepo) Update(
	ctx context.Context,
	id uint,
	editorID uint,
	title *string,
	threadType *string,
) (*gdomain.Thread, error) {
	var thread gdomain.Thread

	err := r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем, существует ли тред
		if err := tx.First(&thread, "id = ?", id).Error; err != nil {
			return err
		}

		// Проверяем права
		if thread.CreatorID != editorID {
			return ErrPermissionDenied
		}

		// Собираем обновляемые поля
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		if title != nil {
			updates["title"] = *title
		}
		if threadType != nil {
			updates["type"] = *threadType
		}

		// Выполняем обновление
		if err := tx.Model(&thread).Updates(updates).Error; err != nil {
			return err
		}

		// Возвращаем актуальные данные
		return tx.First(&thread, "id = ?", id).Error
	})

	if err != nil {
		return nil, err
	}

	return &thread, nil
}

func (r *ThreadRepo) GetThreadMembers(ctx context.Context, threadID uint) ([]gdomain.ThreadUser, error) {
	var members []gdomain.ThreadUser
	if err := r.Db.WithContext(ctx).
		Table("thread_users").
		Where("thread_id = ? AND is_member = ?", threadID, true).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *ThreadRepo) GetAccessibleThreadIDs(ctx context.Context, userID uint) ([]uint, error) {
	var threadIDs []uint
	err := r.Db.WithContext(ctx).
		Table("thread_users").
		Where("user_id = ? AND is_member = ?", userID, true).
		Pluck("thread_id", &threadIDs).Error
	if err != nil {
		return nil, err
	}
	return threadIDs, nil
}

func (r *ThreadRepo) GetAccessibleThreadIDsBySpool(ctx context.Context, userID, spoolID uint) ([]uint, error) {
	var threadIDs []uint

	err := r.Db.WithContext(ctx).
		Table("thread_users tu").
		Select("tu.thread_id").
		Joins("JOIN threads t ON t.id = tu.thread_id").
		Where("tu.user_id = ? AND tu.is_member = ? AND t.spool_id = ?", userID, true, spoolID).
		Pluck("tu.thread_id", &threadIDs).Error

	if err != nil {
		return nil, err
	}

	return threadIDs, nil
}
