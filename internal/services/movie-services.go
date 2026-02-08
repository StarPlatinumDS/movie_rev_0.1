package services

import (
	"context"
	"fmt"
	"io"
	"movie-review/internal/database"
	"movie-review/internal/models"

	"movie-review/internal/storage"

	"github.com/google/uuid"
)

type MovieService struct {
	repo      *database.MovieRepository
	minio     *storage.MinioStorage
	bucket    string
	publicURL string // для формирования ссылок: "http://localhost:9000/my-bucket"
}

func NewMovieService(repo *database.MovieRepository, minio *storage.MinioStorage, bucket string, publicURL string) *MovieService {
	return &MovieService{
		repo:      repo,
		minio:     minio,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

func (s *MovieService) CreateMovie(ctx context.Context, name string, file io.Reader, fileName string) (*models.MovieResponse, error) {
	// generate unique filename
	uniqueName := fmt.Sprintf("%s_%s", uuid.New().String(), fileName)

	// upload to minio
	err := s.minio.UploadFile(ctx, s.bucket, uniqueName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %s", err)
	}

	// save to pg db
	movie := &models.Movie{
		Name:       name,
		PictureKey: uniqueName,
	}

	err = s.repo.Create(ctx, movie)
	if err != nil {
		// if db didn't save -> delete the file
		s.minio.DeleteFile(ctx, s.bucket, uniqueName)
		return nil, fmt.Errorf("failed to upload to DB: %s", err)
	}

	// make a response w/ full picture URL
	pictureURL := fmt.Sprintf("%s/%s", s.publicURL, uniqueName)

	return &models.MovieResponse{
		ID:         movie.ID,
		Name:       movie.Name,
		PictureURL: pictureURL,
	}, nil
}

// GetAllMovies returns to the user all movies with picture links
func (s *MovieService) GetAllMovies(ctx context.Context) ([]models.MovieResponse, error) {
	movies, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []models.MovieResponse
	for _, movie := range movies {
		pictureURL := fmt.Sprintf("%s/%s", s.publicURL, movie.PictureKey)
		responses = append(responses, models.MovieResponse{
			ID:         movie.ID,
			Name:       movie.Name,
			PictureURL: pictureURL,
		})
	}

	return responses, nil
}

// GetMovieByID returns a movie by id and a full picture URL
func (s *MovieService) GetMovieByID(ctx context.Context, id int) (*models.MovieResponse, error) {
	movie, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pictureURL := fmt.Sprintf("%s/%s", s.publicURL, movie.PictureKey)

	return &models.MovieResponse{
		ID:         movie.ID,
		Name:       movie.Name,
		PictureURL: pictureURL,
	}, nil
}

// UpdateMovie returns an update version of stored movie,
// updates everything except createdAt
func (s *MovieService) UpdateMovie(ctx context.Context, id int, name string, file io.Reader, fileName string) (*models.MovieResponse, error) {
	oldMovie, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pictureKey := oldMovie.PictureKey

	// if new file is uploaded - replace an old one
	if file != nil && fileName != "" {
		uniiqueName := fmt.Sprintf("%s_%s", uuid.New().String(), fileName)

		err = s.minio.UploadFile(ctx, s.bucket, uniiqueName, file)
		if err != nil {
			return nil, fmt.Errorf("failed to upload new file: %s", err)
		}

		// delete an old file
		s.minio.DeleteFile(ctx, s.bucket, oldMovie.PictureKey)
		pictureKey = uniiqueName
	}

	// update in pg db
	err = s.repo.Update(ctx, id, name, pictureKey)
	if err != nil {
		return nil, err
	}

	pictureURL := fmt.Sprintf("%s/%s", s.publicURL, pictureKey)

	return &models.MovieResponse{
		ID:         id,
		Name:       name,
		PictureURL: pictureURL,
	}, nil
}

// DeleteMovie deletes from db & storage, returns err if something
// went wrong
func (s *MovieService) DeleteMovie(ctx context.Context, id int) error {
	movie, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// delete from minio
	err = s.minio.DeleteFile(ctx, s.bucket, movie.PictureKey)
	if err != nil {
		return fmt.Errorf("failed to delete file: %s", err)
	}

	// delete from db
	return s.repo.Delete(ctx, id)
}
