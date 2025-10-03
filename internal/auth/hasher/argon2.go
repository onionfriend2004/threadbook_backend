package hasher

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/onionfriend2004/threadbook_backend/config"
)

var (
	ErrEmptyPassword = errors.New("password is empty")
	ErrEmptyHash     = errors.New("hash is empty")
	ErrInvalidParams = errors.New("invalid argon2 parameters")
)

type argon2Hasher struct {
	params *argon2id.Params
}

func NewArgon2HasherFromConfig(cfg config.Config) (*argon2Hasher, error) {
	if cfg.Argon2.Memory == 0 || cfg.Argon2.Iterations == 0 || cfg.Argon2.Parallelism == 0 {
		return nil, ErrInvalidParams
	}

	params := &argon2id.Params{
		Memory:      cfg.Argon2.Memory,
		Iterations:  cfg.Argon2.Iterations,
		Parallelism: cfg.Argon2.Parallelism,
		SaltLength:  cfg.Argon2.SaltLength,
		KeyLength:   cfg.Argon2.KeyLength,
	}

	if params.SaltLength == 0 {
		params.SaltLength = 16
	}
	if params.KeyLength == 0 {
		params.KeyLength = 32
	}

	return &argon2Hasher{params: params}, nil
}

func (h *argon2Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	return argon2id.CreateHash(password, h.params)
}

func (h *argon2Hasher) Verify(password, hash string) (bool, error) {
	if password == "" {
		return false, ErrEmptyPassword
	}
	if hash == "" {
		return false, ErrEmptyHash
	}
	return argon2id.ComparePasswordAndHash(password, hash)
}

var _ HasherInterface = (*argon2Hasher)(nil)
