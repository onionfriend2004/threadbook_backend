package usecase

import "github.com/onionfriend2004/threadbook_backend/internal/auth/external"

type AuthUsecaseInterface struct {
}

type authUsecase struct {
	userRepo external.UserRepoInterface
}
