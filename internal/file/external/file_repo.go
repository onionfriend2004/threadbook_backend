package external

import (
	"context"
	"io"
)

type FileRepoInterface interface {
	GetFile(ctx context.Context, filename string) ([]byte, string, error)
	SaveFile(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, filename string) error
}
