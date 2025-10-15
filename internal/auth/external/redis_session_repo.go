package external

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"github.com/redis/go-redis/v9"
)

type sessionRepo struct {
	redisClient *redis.Client
	sessionTTL  time.Duration
}

func NewSessionRepo(redisClient *redis.Client, sessionTTL time.Duration) SessionRepoInterface {
	return &sessionRepo{
		redisClient: redisClient,
		sessionTTL:  sessionTTL,
	}
}

func (r *sessionRepo) generateSessionID() (string, error) {
	sessionId := uuid.NewString()
	return sessionId, nil
}

func (r *sessionRepo) GetSessionByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := "session_id:" + sessionID
	data, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, ErrInvalidSessionData
	}

	return &session, nil
}

func (r *sessionRepo) AddSessionForUser(ctx context.Context, user *gdomain.User) (*domain.Session, error) {
	sessionID := uuid.NewString()
	key := "session_id:" + sessionID
	now := time.Now()
	expiresAt := now.Add(r.sessionTTL)
	session := &domain.Session{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := r.redisClient.Set(ctx, key, data, r.sessionTTL).Err(); err != nil {
		return nil, err
	}

	return session, nil
}

func (r *sessionRepo) DelSessionByID(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session_id:%s", sessionID)
	return r.redisClient.Del(ctx, key).Err()
}

var _ SessionRepoInterface = (*sessionRepo)(nil)
