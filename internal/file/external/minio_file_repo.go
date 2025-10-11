package external

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type FileRepo struct {
	client     *minio.Client
	bucketName string
}

func NewFileRepo(client *minio.Client, bucket string) *FileRepo {
	return &FileRepo{
		client:     client,
		bucketName: bucket,
	}
}

func (r *FileRepo) GetFile(ctx context.Context, filename string) ([]byte, string, error) {
	obj, err := r.client.GetObject(ctx, r.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrGetObject, err)
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrReadObject, err)
	}

	info, err := r.client.StatObject(ctx, r.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrStatObject, err)
	}

	return buf.Bytes(), info.ContentType, nil
}

func (r *FileRepo) SaveFile(ctx context.Context, filename string, data []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, r.bucketName, filename, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", ErrPutObject, err)
	}
	return nil
}

func (r *FileRepo) DeleteFile(ctx context.Context, filename string) error {
	err := r.client.RemoveObject(ctx, r.bucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", ErrRemoveObject, err)
	}
	return nil
}
