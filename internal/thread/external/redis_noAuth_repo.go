package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/external/domain"
	"github.com/redis/go-redis/v9"
)

type NoAuthRepoInterface interface {
	CreateNoAuthSession(ctx context.Context, threadID uint) (*domain.NoAuthSession, error)
	GetNoAuthSession(ctx context.Context, sessionID string) (*domain.NoAuthSession, error)
	GetThreadSessions(ctx context.Context, threadID uint) ([]*domain.NoAuthSession, error)
	DeleteNoAuthSession(ctx context.Context, sessionID string) error
}

type noAuthRepo struct {
	redisClient *redis.Client
}

func NewNoAuthRepo(redisClient *redis.Client) NoAuthRepoInterface {
	return &noAuthRepo{
		redisClient: redisClient,
	}
}

func (r *noAuthRepo) CreateNoAuthSession(ctx context.Context, threadID uint) (*domain.NoAuthSession, error) {
	sessionID := uuid.NewString()
	key := "NoAuth_id:" + sessionID
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	session := &domain.NoAuthSession{
		ID:        sessionID,
		ThreadID:  threadID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	// Транзакция чтобы обе записи создавались атомарно
	pipe := r.redisClient.TxPipeline()

	// Основная запись: session_id -> session_data
	pipe.Set(ctx, key, data, 24*time.Hour)

	// Обратный индекс: thread_id -> [session_ids]
	threadKey := fmt.Sprintf("thread_sessions:%d", threadID)
	pipe.SAdd(ctx, threadKey, sessionID)
	pipe.Expire(ctx, threadKey, 24*time.Hour) // На всякий случай TTL

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Получить сессию по ID
func (r *noAuthRepo) GetNoAuthSession(ctx context.Context, sessionID string) (*domain.NoAuthSession, error) {
	key := "NoAuth_id:" + sessionID
	data, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var session domain.NoAuthSession
	err = json.Unmarshal(data, &session)
	return &session, err
}

// Получить все сессии для треда
func (r *noAuthRepo) GetThreadSessions(ctx context.Context, threadID uint) ([]*domain.NoAuthSession, error) {
	threadKey := fmt.Sprintf("thread_sessions:%d", threadID)

	// Получаем все session_id для этого треда
	sessionIDs, err := r.redisClient.SMembers(ctx, threadKey).Result()
	if err != nil {
		return nil, err
	}

	// Параллельно получаем данные всех сессий
	pipe := r.redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(sessionIDs))

	for i, sessionID := range sessionIDs {
		key := "NoAuth_id:" + sessionID
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Декодируем результаты
	var sessions []*domain.NoAuthSession
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue // Пропускаем просроченные
		}

		var session domain.NoAuthSession
		if json.Unmarshal(data, &session) == nil {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

// Удалить сессию
func (r *noAuthRepo) DeleteNoAuthSession(ctx context.Context, sessionID string) error {
	// Сначала получаем сессию чтобы узнать thread_id
	session, err := r.GetNoAuthSession(ctx, sessionID)
	if err != nil {
		return err
	}

	pipe := r.redisClient.TxPipeline()

	// Удаляем основную запись
	pipe.Del(ctx, "NoAuth_id:"+sessionID)

	// Удаляем из индекса
	threadKey := fmt.Sprintf("thread_sessions:%d", session.ThreadID)
	pipe.SRem(ctx, threadKey, sessionID)

	_, err = pipe.Exec(ctx)
	return err
}
