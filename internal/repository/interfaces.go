package repository

import (
	"context"
	"io"
	"movie-review/internal/models"
)

// interface for MovieRepository repo
type MovieRepository interface {
	Create(ctx context.Context, movie *models.Movie) error
	GetAll(ctx context.Context) ([]models.Movie, error)
	GetByID(ctx context.Context, id int) (*models.Movie, error)
	SearchMovies(ctx context.Context, query string, yearFrom int, yearTo int, genres []string, hasComment *bool, offset, limit int) ([]models.Movie, error)
	GetAllGenres(ctx context.Context) ([]string, error)
	Update(ctx context.Context, movie *models.Movie) error
	Delete(ctx context.Context, id int) error
}

// FileStorage is interface for MinioStorage repo
type FileStorage interface {
	UploadFile(ctx context.Context, bucket string, objectName string, file io.Reader) error
	DeleteFile(ctx context.Context, bucket, objectName string) error
}
