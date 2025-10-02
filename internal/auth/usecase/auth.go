package usecase

import (
	"context"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/external"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/hasher"
)

type AuthUsecaseInterface struct {
	SignUpUser(ctx context.Context, user *domain.User) (*domain.User, error)
	SignInUser(ctx context.Context, user *domain.User) (*domain.User, error)
	SignOutUser(ctx context.Context, session *domain.Session) (error)
	AuthenticateUser(ctx context.Context, session *domain.Session) (*domain.User, error)
}

type authUsecase struct {
	userRepo    external.UserRepoInterface
	sessionRepo external.SessionRepoInterface
	hasher      hasher.HasherInterface
}

func NewAuthService(userRepo external.UserRepoInterface, sessionRepo external.SessionRepoInterface, hasher hasher.HasherInterface) AuthUsecaseInterface {
	return &authUsecase{userRepo: userRepo, sessionRepo: sessionRepo, hasher: hasher}
}



var _ AuthUsecaseInterface = (*authUsecase)(nil)
