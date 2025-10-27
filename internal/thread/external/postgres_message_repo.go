package external

import (
	"context"
	"fmt"

	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"gorm.io/gorm"
)

type messageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) MessageRepoInterface {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(ctx context.Context, m *gdomain.Message) error {
	if m == nil {
		return fmt.Errorf("message is nil")
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepo) CreateWithPayloads(ctx context.Context, m *gdomain.Message) error {
	if m == nil {
		return fmt.Errorf("message is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		// Если payloads есть — установить MessageID и вставить
		if len(m.Payloads) > 0 {
			for i := range m.Payloads {
				m.Payloads[i].MessageID = m.ID
			}
			if err := tx.Create(&m.Payloads).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *messageRepo) GetByThreadID(ctx context.Context, threadID uint, limit, offset int) ([]gdomain.Message, error) {
	var msgs []gdomain.Message
	q := r.db.WithContext(ctx).
		Preload("User").
		Preload("Payloads").
		Where("thread_id = ?", threadID).
		Order("created_at ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	if err := q.Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *messageRepo) GetByID(ctx context.Context, id uint) (*gdomain.Message, error) {
	var m gdomain.Message
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Payloads").
		First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *messageRepo) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&gdomain.Message{}, id).Error
}

func (r *messageRepo) CountByThreadID(ctx context.Context, threadID uint) (int64, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&gdomain.Message{}).Where("thread_id = ?", threadID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}
