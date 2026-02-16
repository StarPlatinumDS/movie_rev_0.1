package repository

import (
	"context"
	"io"
	"movie-review/internal/models"
)

// interface for movieRepository repo
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

// MovieService is an interface for movieService
type MovieService interface {
	CreateMovie(ctx context.Context, name string, file io.Reader, fileName string,
		genres []string, year int, description string, rating int, worldRating float32, comment string) (*models.MovieResponse, error)
	GetAllMovies(ctx context.Context) ([]models.MovieResponse, error)
	GetMovieByID(ctx context.Context, id int) (*models.MovieResponse, error)
	SearchMovies(ctx context.Context, query string, yearFrom, yearTo int, genres []string, hasComment *bool, offset, limit int) ([]models.MovieResponse, error)
	GetAllGenres(ctx context.Context) ([]string, error)
	UpdateMovie(ctx context.Context, id int, name string, file io.Reader, fileName string,
		genres []string, year int, description string, rating int, worldRating float32, comment string) (*models.MovieResponse, error)
	DeleteMovie(ctx context.Context, id int) error
}
