package external

import (
	"context"
	"errors"
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
	data, err := r.redisClient.Get(ctx, sessionID).Bytes()
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
	now := time.Now()
	expiresAt := now.Add(r.sessionTTL)
	session := &domain.Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := r.redisClient.Set(ctx, sessionID, data, r.sessionTTL).Err(); err != nil {
		return nil, err
	}

	return session, nil
}

func (r *sessionRepo) DelSessionByID(ctx context.Context, sessionID string) error {
	return r.redisClient.Del(ctx, sessionID).Err()
}

var _ SessionRepoInterface = (*sessionRepo)(nil)
