package usecase

import (
	"context"
	"errors"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
	"github.com/onionfriend2004/threadbook_backend/internal/gdomain"
	"go.uber.org/zap"
)

type AuthUsecaseInterface interface {
	SignUpUser(ctx context.Context, input SignUpInput) (*gdomain.User, error)
	SignInUser(ctx context.Context, input SignInInput) (*gdomain.User, error)
	SignOutUser(ctx context.Context, sessionID string) error
	AuthenticateUser(ctx context.Context, sessionID string) (*gdomain.User, error)
	CreateSessionForUser(ctx context.Context, user *gdomain.User) (*domain.Session, error)

	VerifyUserEmail(ctx context.Context, userID int, code int) error
	ResendVerifyCode(ctx context.Context, userID int) error
}

type authUsecase struct {
	userRepo            external.UserRepoInterface
	sessionRepo         external.SessionRepoInterface
	sendCodeRepo        external.SendCodeRepoInterface
	verifyCodeRepo      external.VerifyCodeRepoInterface
	hasher              hasher.HasherInterface
	logger              *zap.Logger
	attemptSendCodeRepo external.AttemptSendCodeRepoInterface
	maxResendAttempts   int
}

func NewAuthUsecase(
	userRepo external.UserRepoInterface,
	sessionRepo external.SessionRepoInterface,
	sendCodeRepo external.SendCodeRepoInterface,
	verifyCodeRepo external.VerifyCodeRepoInterface,
	hasher hasher.HasherInterface,
	logger *zap.Logger,
	attemptSendCodeRepo external.AttemptSendCodeRepoInterface,
	maxResendAttempts int,
) AuthUsecaseInterface {
	return &authUsecase{
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		sendCodeRepo:        sendCodeRepo,
		verifyCodeRepo:      verifyCodeRepo,
		hasher:              hasher,
		logger:              logger,
		attemptSendCodeRepo: attemptSendCodeRepo,
		maxResendAttempts:   maxResendAttempts,
	}
}

func (u *authUsecase) SignUpUser(ctx context.Context, input SignUpInput) (*gdomain.User, error) {
	if input.Email == "" || input.Username == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	email := gdomain.NormalizeEmail(input.Email)
	username := gdomain.NormalizeUsername(input.Username)

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

	newUser := gdomain.User{
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
	}

	createdUser, err := u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		u.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	verifyCode, err := u.verifyCodeRepo.GenerateCode()
	if err != nil {
		u.logger.Error("failed to generate verify code", zap.Error(err))
		return createdUser, nil
	}

	err = u.verifyCodeRepo.SaveCode(ctx, createdUser.ID, verifyCode)
	if err != nil {
		u.logger.Error("failed to save verify code", zap.Error(err))
		return createdUser, nil
	}

	err = u.sendCodeRepo.SendVerifyCodeForUser(verifyCode, createdUser)
	if err != nil {
		u.logger.Error("failed to send verify code in broker", zap.Error(err))
		return createdUser, nil
	}

	return createdUser, nil
}

func (u *authUsecase) SignInUser(ctx context.Context, input SignInInput) (*gdomain.User, error) {
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	email := gdomain.NormalizeEmail(input.Email)
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

func (u *authUsecase) AuthenticateUser(ctx context.Context, sessionID string) (*gdomain.User, error) {
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

func (u *authUsecase) CreateSessionForUser(ctx context.Context, user *gdomain.User) (*domain.Session, error) {
	if user == nil {
		return nil, ErrInvalidInput
	}
	return u.sessionRepo.AddSessionForUser(ctx, user)
}

func (u *authUsecase) VerifyUserEmail(ctx context.Context, userID int, code int) error {
	if userID <= 0 || code < 100000 || code > 999999 {
		return ErrInvalidInput
	}
	valid, err := u.verifyCodeRepo.VerifyCode(ctx, uint(userID), code)
	if err != nil {
		u.logger.Error("failed to verify code",
			zap.Int("user_id", userID),
			zap.Int("code", code),
			zap.Error(err))
		return err
	}
	if !valid {
		return ErrCodeIncorrect
	}

	if err := u.userRepo.VerifyUserEmail(ctx, uint(userID)); err != nil {
		u.logger.Error("failed to verify user email in DB",
			zap.Int("user_id", userID),
			zap.Error(err))
		return err
	}

	if err := u.attemptSendCodeRepo.ResetSendAttempts(ctx, uint(userID)); err != nil {
		u.logger.Warn("failed to reset send attempts after successful verification",
			zap.Int("user_id", userID),
			zap.Error(err))
		// Не прерываем — верификация успешна, сброс вторичен
	}

	u.logger.Info("user email verified successfully", zap.Int("user_id", userID))
	return nil
}

func (u *authUsecase) ResendVerifyCode(ctx context.Context, userID int) error {
	if userID <= 0 {
		return ErrInvalidInput
	}

	// Проверяем, существует ли пользователь и не подтверждён ли уже email
	user, err := u.userRepo.GetUserByID(ctx, uint(userID))
	if err != nil {
		if errors.Is(err, external.ErrUserNotFound) {
			return ErrUserNotFound
		}
		u.logger.Error("failed to get user by ID during resend", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	if user.EmailVerify {
		// Email уже подтверждён — переотправка не нужна
		return ErrAlreadyConfirmed
	}

	// Проверяем количество попыток отправки
	attempts, err := u.attemptSendCodeRepo.GetSendAttempts(ctx, uint(userID))
	if err != nil {
		u.logger.Error("failed to get send attempts", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	if attempts >= u.maxResendAttempts {
		return ErrTooManyAttempts
	}

	// Генерируем новый код
	newCode, err := u.verifyCodeRepo.GenerateCode()
	if err != nil {
		u.logger.Error("failed to generate new verification code", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	// Сохраняем новый код (старый перезаписывается)
	if err := u.verifyCodeRepo.SaveCode(ctx, uint(userID), newCode); err != nil {
		u.logger.Error("failed to save new verification code", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	// Отправляем код (в очередь)
	if err := u.sendCodeRepo.SendVerifyCodeForUser(newCode, user); err != nil {
		u.logger.Error("failed to send verification code via broker", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	// Увеличиваем счётчик попыток отправки
	if err := u.attemptSendCodeRepo.IncrementSendAttempts(ctx, uint(userID)); err != nil {
		u.logger.Warn("failed to increment send attempts", zap.Error(err), zap.Int("user_id", userID))
		// Не прерываем — код отправлен, счётчик вторичен
	}

	u.logger.Info("verification code resent successfully",
		zap.Int("user_id", userID),
		zap.Int("attempts", attempts+1),
		zap.Int("max_attempts", u.maxResendAttempts))

	return nil
}

var _ AuthUsecaseInterface = (*authUsecase)(nil)
