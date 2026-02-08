package services

import (
	"movie-review/internal/database"

	"movie-review/internal/storage"
)

type MovieService struct {
	repo      *database.MovieRepository
	minio     *storage.MinioStorage
	bucket    string
	publicURL string // для формирования ссылок: "http://localhost:9000/my-bucket"
}

func NewMovieService()
