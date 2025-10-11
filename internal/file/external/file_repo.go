package external

import "context"

type FileRepoInterface interface {
	GetFile(ctx context.Context, filename string) ([]byte, string, error)
	SaveFile(ctx context.Context, filename string, data []byte, contentType string) error
	DeleteFile(ctx context.Context, filename string) error
}
