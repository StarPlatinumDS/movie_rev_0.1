package storage

import (
	"context"
	"io"
	"movie-review/internal/repository"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) repository.FileStorage {
	return &MinioStorage{client: client}
}

func (s *MinioStorage) UploadFile(ctx context.Context, bucket string, objectName string, file io.Reader) error {
	_, err := s.client.PutObject(ctx, bucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: "image/jpeg", // need png, webp
	})
	return err
}

func (s *MinioStorage) DeleteFile(ctx context.Context, bucket, objectName string) error {
	return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}
