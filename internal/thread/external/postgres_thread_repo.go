package external

import (
	"context"
	"errors"
	"time"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			return ErrCreateThread
		}

		if count == 0 {
			return ErrUserNotInSpool
		}

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
			return ErrCreateThread
		}

		if threadType == "public" {
			var userIDs []int
			if err := tx.
				Table("user_spools").
				Select("user_id").
				Where("spool_id = ?", spoolID).
				Pluck("user_id", &userIDs).Error; err != nil {
				return ErrCreateThread
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
					return ErrCreateThread
				}
			}
		} else {
			if err := tx.Table("thread_users").Create(map[string]interface{}{
				"user_id":   creatorID,
				"thread_id": thread.ID,
			}).Error; err != nil {
				return ErrCreateThread
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

	err := r.Db.
		Table("threads AS t").
		Joins("JOIN thread_users ut ON ut.thread_id = t.id").
		Where("t.spool_id = ? AND ut.user_id = ?", spoolID, userID).
		Find(&threads).Error

	if err != nil {
		return nil, ErrGetThreads
	}
	return threads, nil
}

func (r *ThreadRepo) CloseThread(id uint, userID uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread
	if err := r.Db.First(&thread, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, ErrCloseThread
	}
	if thread.CreatorID == userID {
		thread.IsClosed = true
		if err := r.Db.Save(&thread).Error; err != nil {
			return nil, ErrCloseThread
		}
		return &thread, nil
	}
	return nil, ErrUserNoAccess
}

func (r *ThreadRepo) GetThreadByID(ctx context.Context, threadID uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread
	if err := r.Db.WithContext(ctx).First(&thread, threadID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, ErrThreadNotFound
	}
	return &thread, nil
}

func (r *ThreadRepo) CheckRightsUserOnThreadRoom(ctx context.Context, threadID uint, userID uint) (bool, error) {
	var count int64
	err := r.Db.WithContext(ctx).
		Table("thread_users").
		Where("user_id = ? AND thread_id = ? AND is_member = ?", userID, threadID, true).
		Count(&count).Error
	if err != nil {
		return false, ErrRightsCheck
	}
	return count > 0, nil
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
		if err := tx.First(&thread, "id = ?", id).Error; err != nil {
			return ErrThreadNotFound
		}

		if thread.CreatorID != editorID {
			return ErrPermissionDenied
		}

		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		if title != nil {
			updates["title"] = *title
		}
		if threadType != nil {
			updates["type"] = *threadType
		}

		if err := tx.Model(&thread).Updates(updates).Error; err != nil {
			return ErrUpdateThread
		}

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
		return nil, ErrGetMembers
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
		return nil, ErrGetAccessibleIDs
	}
	return threadIDs, nil
}

func (r *ThreadRepo) InviteToThread(ctx context.Context, inviterID uint, inviteeUsernames []string, threadID uint) error {
	return r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var thread gdomain.Thread
		if err := tx.First(&thread, threadID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrThreadNotFound
			}
			return ErrInviteFailed
		}

		if thread.Type != "private" {
			return ErrInvalidThreadType
		}

		if inviterID != thread.CreatorID {
			return ErrUserNoAccess
		}

		for _, username := range inviteeUsernames {
			var invitee gdomain.User
			if err := tx.Where("username = ?", username).First(&invitee).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrUserNotFound
				}
				return ErrInviteFailed
			}

			var inSpool int64
			if err := tx.Table("user_spools").
				Where("user_id = ? AND spool_id = ?", invitee.ID, thread.SpoolID).
				Count(&inSpool).Error; err != nil {
				return ErrInviteFailed
			}
			if inSpool == 0 {
				return ErrUserNotInSpool
			}

			threadUser := gdomain.ThreadUser{
				UserID:   invitee.ID,
				ThreadID: thread.ID,
				IsMember: true,
			}

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&threadUser).Error; err != nil {
				return ErrInviteFailed
			}
		}

		return nil
	})
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
		return nil, ErrGetAccessibleIDs
	}

	return threadIDs, nil
}
