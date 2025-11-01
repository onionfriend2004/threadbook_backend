package external

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type FileRepo struct {
	client     *minio.Client
	BucketName string
}

func NewFileRepo(client *minio.Client, bucket string) FileRepoInterface {

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Printf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			log.Printf("failed to create bucket %q: %v", bucket, err)
		} else {
			log.Printf("bucket %q successfully created", bucket)
		}
	} else {
		log.Printf("bucket %q already exists", bucket)
	}

	return &FileRepo{
		client:     client,
		BucketName: bucket,
	}
}

func (r *FileRepo) GetFile(ctx context.Context, bucket, filename string) ([]byte, string, error) {
	obj, err := r.client.GetObject(ctx, bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrGetObject, err)
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrReadObject, err)
	}

	info, err := r.client.StatObject(ctx, r.BucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", ErrStatObject, err)
	}

	return buf.Bytes(), info.ContentType, nil
}

func (r *FileRepo) SaveFile(ctx context.Context, originalName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Извлекаем расширение файла
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".jpg" // дефолтное расширение
	}

	newFilename := fmt.Sprintf("%s%s", uuid.NewString(), ext)

	_, err := r.client.PutObject(ctx, r.BucketName, newFilename, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrPutObject, err)
	}

	return newFilename, nil
}

func (r *FileRepo) DeleteFile(ctx context.Context, filename string) error {
	err := r.client.RemoveObject(ctx, r.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", ErrRemoveObject, err)
	}
	return nil
}

func (r *FileRepo) GetBucketName() string {
	return r.BucketName
}
