package external

import (
	"context"
	"errors"
	"fmt"
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

func (r *ThreadRepo) Create(ctx context.Context, creatorID uint, spoolID *uint, title string, threadType gdomain.ThreadType, accessLevel *uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread

	err := r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trueAccessLevel uint
		if accessLevel != nil {
			trueAccessLevel = *accessLevel
		} else {
			trueAccessLevel = 0
		}
		thread = gdomain.Thread{
			CreatorID:   creatorID,
			SpoolID:     spoolID,
			Title:       title,
			Type:        threadType,
			AccessLevel: trueAccessLevel,
			IsClosed:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := tx.Create(&thread).Error; err != nil {
			return ErrCreateThread
		}

		// Добавляем создателя в thread_users только для приватных тредов
		if thread.Type == gdomain.ThreadTypePrivate {
			threadUser := gdomain.ThreadUser{
				UserID:   creatorID,
				ThreadID: thread.ID,
				Status:   gdomain.ThreadUserStatusActive,
				JoinedAt: time.Now(),
			}

			if err := tx.Create(&threadUser).Error; err != nil {
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

func (r *ThreadRepo) IsUserThreadMember(ctx context.Context, userID uint, threadID uint) (bool, error) {
	var count int64

	// Дебаг параметров
	fmt.Printf("Параметры запроса: threadID=%d, userID=%d, status='%s'\n",
		threadID, userID, string(gdomain.ThreadUserStatusActive))

	err := r.Db.WithContext(ctx).
		Table("thread_users").
		Where("thread_id = ? AND user_id = ? AND status = ?", threadID, userID, gdomain.ThreadUserStatusActive).
		Count(&count).Error

	fmt.Printf("Результат запроса: count=%d, err=%v\n", count, err)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ThreadRepo) GetBySpoolID(ctx context.Context, userID, spoolID uint) ([]*gdomain.Thread, error) {
	var threads []*gdomain.Thread

	err := r.Db.WithContext(ctx).
		Raw(`
            SELECT t.* FROM threads t
            LEFT JOIN user_spools us ON us.user_id = ? AND us.spool_id = ?
            LEFT JOIN thread_users tu ON tu.thread_id = t.id AND tu.user_id = ? AND tu.status = 'active'
            WHERE t.spool_id = ? AND (
                (t.type = 'public' AND us.access_level >= t.access_level) OR
                (t.type = 'private' AND tu.user_id IS NOT NULL)
            )
        `, userID, spoolID, userID, spoolID).
		Find(&threads).Error

	if err != nil {
		return nil, ErrGetThreads
	}
	return threads, nil
}

func (r *ThreadRepo) CloseThread(ctx context.Context, id uint) (*gdomain.Thread, error) {
	var thread gdomain.Thread
	if err := r.Db.WithContext(ctx).First(&thread, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, ErrCloseThread
	}

	thread.IsClosed = true
	thread.UpdatedAt = time.Now()

	if err := r.Db.WithContext(ctx).Save(&thread).Error; err != nil {
		return nil, ErrCloseThread
	}
	return &thread, nil
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
	title *string,
	threadType *string,
	accessLevel *uint,
) (*gdomain.Thread, error) {
	var thread gdomain.Thread

	err := r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&thread, "id = ?", id).Error; err != nil {
			return ErrThreadNotFound
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
		if accessLevel != nil {
			updates["access_level"] = *accessLevel
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

func (r *ThreadRepo) GetUserThreadAccessLevelsToUpdate(ctx context.Context, userID uint, threadID uint) (userAccessLevel uint, threadAccessLevel uint, err error) {
	var result struct {
		ThreadType        string
		ThreadAccessLevel uint
		SpoolID           *uint
	}

	// Сначала получаем информацию о треде
	err = r.Db.WithContext(ctx).
		Table("threads").
		Select("type, access_level, spool_id").
		Where("id = ?", threadID).
		Scan(&result).Error

	if err != nil {
		return 0, 0, err
	}

	// Для приватных тредов проверяем участников
	if result.ThreadType == "private" {
		// Проверяем, является ли пользователь участником приватного треда
		var isMember bool
		err = r.Db.WithContext(ctx).
			Table("thread_users").
			Select("COUNT(*) > 0").
			Where("thread_id = ? AND user_id = ? AND status = ?", threadID, userID, gdomain.ThreadUserStatusActive).
			Scan(&isMember).Error
		if err != nil {
			return 0, 0, err
		}

		if isMember {
			// Участник приватного треда имеет доступ
			return 1, 0, nil // userLevel > threadLevel = доступ разрешен
		} else {
			// Не участник - доступ запрещен
			return 0, 1, nil // userLevel < threadLevel = доступ запрещен
		}
	}

	// Для публичных тредов проверяем уровень доступа
	if result.SpoolID == nil {
		// Если тред не привязан к спулу - доступ запрещен
		return 0, 1, nil
	}

	// Получаем уровень доступа пользователя в спуле
	var userSpool gdomain.UserSpool
	err = r.Db.WithContext(ctx).
		Where("user_id = ? AND spool_id = ?", userID, *result.SpoolID).
		First(&userSpool).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Пользователь не состоит в спуле
			return 0, result.ThreadAccessLevel, nil
		}
		return 0, 0, err
	}

	return userSpool.AccessLevel, result.ThreadAccessLevel, nil
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
		// // Проверяем что тред приватный и приглашающий - участник
		// var thread gdomain.Thread
		// if err := tx.First(&thread, threadID).Error; err != nil {
		// 	return ErrThreadNotFound
		// }

		// if thread.Type != "private" {
		// 	return ErrInvalidThreadType
		// }

		// // Проверяем что приглашающий в треде
		// var count int64
		// if err := tx.Table("thread_users").
		// 	Where("thread_id = ? AND user_id = ? AND status = ?", threadID, inviterID, gdomain.ThreadUserStatusActive).
		// 	Count(&count).Error; err != nil || count == 0 {
		// 	return ErrUserNoAccess
		// }

		for _, username := range inviteeUsernames {
			var invitee gdomain.User
			if err := tx.Where("username = ?", username).First(&invitee).Error; err != nil {
				return ErrUserNotFound
			}

			// Обновляем или создаем запись с статусом Active
			threadUser := gdomain.ThreadUser{
				UserID:   invitee.ID,
				ThreadID: threadID,
				Status:   gdomain.ThreadUserStatusActive,
				JoinedAt: time.Now(),
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "joined_at"}),
			}).Create(&threadUser).Error; err != nil {
				return ErrInviteFailed
			}
		}

		return nil
	})
}

// не используется
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

func (r *ThreadRepo) IsThreadOwner(ctx context.Context, userID uint, threadID uint) (bool, error) {
	var thread gdomain.Thread
	err := r.Db.WithContext(ctx).
		Select("creator_id").
		Where("id = ?", threadID).
		First(&thread).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("thread not found")
		}
		return false, err
	}

	return thread.CreatorID == userID, nil
}

func (r *ThreadRepo) AddUserToThread(ctx context.Context, userID, threadID uint) error {
	var thread gdomain.Thread
	if err := r.Db.WithContext(ctx).
		Select("spool_id, type").
		Where("id = ?", threadID).
		First(&thread).Error; err != nil {
		return err
	}

	// Если тред привязан к спулу, проверяем что юзер в спуле
	if thread.SpoolID != nil {
		var count int64
		if err := r.Db.WithContext(ctx).
			Table("user_spools").
			Where("user_id = ? AND spool_id = ? AND is_deleted = false", userID, *thread.SpoolID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrUserNotInSpool
		}
	}

	// Добавляем пользователя в тред
	threadUser := gdomain.ThreadUser{
		UserID:   userID,
		ThreadID: threadID,
		Status:   gdomain.ThreadUserStatusActive,
		JoinedAt: time.Now(),
	}

	return r.Db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "joined_at"}),
		}).
		Create(&threadUser).Error
}

func (r *ThreadRepo) GetPrivateThreadUsers(ctx context.Context, threadID uint) ([]gdomain.User, error) {
	var users []gdomain.User

	err := r.Db.WithContext(ctx).
		Preload("Profile").
		Joins("JOIN thread_users ON users.id = thread_users.user_id").
		Where("thread_users.thread_id = ? AND thread_users.status = ?", threadID, gdomain.ThreadUserStatusActive).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *ThreadRepo) GetUsersWithAccess(ctx context.Context, spoolID uint, minAccessLevel uint) ([]gdomain.User, error) {
	var users []gdomain.User
	if err := r.Db.WithContext(ctx).
		Joins("INNER JOIN user_spools ON users.id = user_spools.user_id").
		Where("user_spools.spool_id = ? AND user_spools.access_level >= ?", spoolID, minAccessLevel).
		Find(&users).Error; err != nil {
		return nil, ErrGetMembers
	}
	return users, nil
}
