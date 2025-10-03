package usecase

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
	"go.uber.org/zap"
)

type AuthUsecaseInterface interface {
	SignUpUser(ctx context.Context, input SignUpInput) (*domain.User, error)
	SignInUser(ctx context.Context, input SignInInput) (*domain.User, error)
	SignOutUser(ctx context.Context, sessionID string) error
	AuthenticateUser(ctx context.Context, sessionID string) (*domain.User, error)
	CreateSessionForUser(ctx context.Context, user *domain.User) (*domain.Session, error)
}

type authUsecase struct {
	userRepo    external.UserRepoInterface
	sessionRepo external.SessionRepoInterface
	hasher      hasher.HasherInterface
	logger      *zap.Logger
}

func NewAuthUsecase(
	userRepo external.UserRepoInterface,
	sessionRepo external.SessionRepoInterface,
	hasher hasher.HasherInterface,
	logger *zap.Logger,
) AuthUsecaseInterface {
	return &authUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		logger:      logger,
	}
}

func (u *authUsecase) SignUpUser(ctx context.Context, input SignUpInput) (*domain.User, error) {
	if input.Email == "" || input.Username == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	email := domain.NormalizeEmail(input.Email)
	username := domain.NormalizeUsername(input.Username)

	emailExists, err := u.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		u.logger.Error("failed to check email existence", zap.Error(err), zap.String("email", email))
		return nil, err
	}
	if emailExists {
		return nil, ErrUserAlreadyExists
	}

	usernameExists, err := u.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		u.logger.Error("failed to check username existence", zap.Error(err), zap.String("username", username))
		return nil, err
	}
	if usernameExists {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := u.hasher.Hash(input.Password)
	if err != nil {
		u.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	newUser := domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
	}

	createdUser, err := u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		u.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	return createdUser, nil
}

func (u *authUsecase) SignInUser(ctx context.Context, input SignInInput) (*domain.User, error) {
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	email := domain.NormalizeEmail(input.Email)
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, external.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		u.logger.Error("failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	valid, err := u.hasher.Verify(input.Password, existingUser.PasswordHash)
	if err != nil {
		u.logger.Error("failed to verify password", zap.Error(err))
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidCredentials
	}

	return existingUser, nil
}

func (u *authUsecase) SignOutUser(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidInput
	}

	if err := u.sessionRepo.DelSessionByID(ctx, sessionID); err != nil {
		u.logger.Error("failed to delete session", zap.Error(err), zap.String("session_id", sessionID))
		return err
	}

	return nil
}

func (u *authUsecase) AuthenticateUser(ctx context.Context, sessionID string) (*domain.User, error) {
	if sessionID == "" {
		return nil, ErrInvalidInput
	}

	storedSession, err := u.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, external.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		u.logger.Error("failed to get session", zap.Error(err), zap.String("session_id", sessionID))
		return nil, err
	}

	user, err := u.userRepo.GetUserByID(ctx, storedSession.UserID)
	if err != nil {
		if errors.Is(err, external.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		u.logger.Error("failed to get user by ID", zap.Error(err), zap.Uint("user_id", storedSession.UserID))
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) CreateSessionForUser(ctx context.Context, user *domain.User) (*domain.Session, error) {
	if user == nil {
		return nil, ErrInvalidInput
	}
	return u.sessionRepo.AddSessionForUser(ctx, user)
}

var _ AuthUsecaseInterface = (*authUsecase)(nil)
