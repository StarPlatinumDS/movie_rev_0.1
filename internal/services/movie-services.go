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

func (s *MovieService) CreateMovie(ctx context.Context, name string, file io.Reader, fileName string,
	genres []string, year int, description string, rating int, worldRating float32, comment string) (*models.MovieResponse, error) {
	// generate unique filename
	uniqueName := fmt.Sprintf("%s_%s", uuid.New().String(), fileName)

	// upload to minio
	err := s.minio.UploadFile(ctx, s.bucket, uniqueName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %s", err)
	}

	// save to pg db
	movie := &models.Movie{
		Name:        name,
		PictureKey:  uniqueName,
		Genres:      genres,
		Year:        year,
		Description: description,
		Rating:      rating,
		WorldRating: worldRating,
		Comment:     comment,
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
		ID:          movie.ID,
		Name:        movie.Name,
		PictureURL:  pictureURL,
		Genres:      movie.Genres,
		Year:        movie.Year,
		Description: movie.Description,
		Rating:      movie.Rating,
		WorldRating: movie.WorldRating,
		Comment:     movie.Comment,
		HasComment:  comment != "",
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
			ID:          movie.ID,
			Name:        movie.Name,
			PictureURL:  pictureURL,
			Genres:      movie.Genres,
			Year:        movie.Year,
			Description: movie.Description,
			Rating:      movie.Rating,
			WorldRating: movie.WorldRating,
			Comment:     movie.Comment,
			HasComment:  movie.Comment != "",
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
		ID:          movie.ID,
		Name:        movie.Name,
		PictureURL:  pictureURL,
		Genres:      movie.Genres,
		Year:        movie.Year,
		Description: movie.Description,
		Rating:      movie.Rating,
		WorldRating: movie.WorldRating,
		Comment:     movie.Comment,
		HasComment:  movie.Comment != "",
	}, nil
}

// SearchMovies returns searched movies with/without filtration
func (s *MovieService) SearchMovies(ctx context.Context, query string, yearFrom, yearTo int, genres []string, hasComment *bool, offset, limit int) ([]models.MovieResponse, error) {
	movies, err := s.repo.SearchMovies(ctx, query, yearFrom, yearTo, genres, hasComment, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %s", err)
	}

	responses := make([]models.MovieResponse, len(movies))
	for i, movie := range movies {
		responses[i] = s.movieToResponse(movie)
	}

	return responses, nil
}

// Returns all genres for the search
func (s *MovieService) GetAllGenres(ctx context.Context) ([]string, error) {
	genres, err := s.repo.GetAllGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres: %s", err)
	}
	return genres, nil
}

// UpdateMovie returns an update version of stored movie,
// updates everything except createdAt
func (s *MovieService) UpdateMovie(ctx context.Context, id int, name string, file io.Reader, fileName string,
	genres []string, year int, description string, rating int, worldRating float32, comment string) (*models.MovieResponse, error) {
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

	updatedMovie := &models.Movie{
		ID:          id,
		Name:        name,
		PictureKey:  pictureKey,
		Genres:      genres,
		Year:        year,
		Description: description,
		Rating:      rating,
		WorldRating: worldRating,
		Comment:     comment,
	}

	// update in pg db
	err = s.repo.Update(ctx, updatedMovie)
	if err != nil {
		return nil, err
	}

	pictureURL := fmt.Sprintf("%s/%s", s.publicURL, pictureKey)

	return &models.MovieResponse{
		ID:          id,
		Name:        name,
		PictureURL:  pictureURL,
		Genres:      genres,
		Year:        year,
		Description: description,
		Rating:      rating,
		WorldRating: worldRating,
		Comment:     comment,
		HasComment:  comment != "",
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

// movieToResponse is a helper func to send movie to the client
func (s *MovieService) movieToResponse(movie models.Movie) models.MovieResponse {
	pictureURL := fmt.Sprintf("%s/%s", s.publicURL, movie.PictureKey)
	return models.MovieResponse{
		ID:          movie.ID,
		Name:        movie.Name,
		PictureURL:  pictureURL,
		Genres:      movie.Genres,
		Year:        movie.Year,
		Description: movie.Description,
		Rating:      movie.Rating,
		WorldRating: movie.WorldRating,
		Comment:     movie.Comment,
		HasComment:  movie.Comment != "",
	}
}
