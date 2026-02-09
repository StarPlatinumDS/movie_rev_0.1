package database

import (
	"context"
	"errors"
	"fmt"
	"movie-review/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository struct {
	pool *pgxpool.Pool
}

func NewMovieRepository(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{pool: pool}
}

// Create saves uploaded movie to our db
func (r *MovieRepository) Create(ctx context.Context, movie *models.Movie) error {
	query := `INSERT INTO movies (name, picture_key) VALUES ($1, $2) RETURNING id, created_at`
	// ????
	err := r.pool.QueryRow(ctx, query, movie.Name, movie.PictureKey).Scan(&movie.ID, &movie.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert movie: %s", err)
	}
	return nil
}

// GetAll returns a slice of all uploaded movies
func (r *MovieRepository) GetAll(ctx context.Context) ([]models.Movie, error) {
	query := `SELECT id, name, picture_key, created_at FROM movies ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load movies: %s", err)
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(&movie.ID, &movie.Name, &movie.PictureKey, &movie.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to load movie: %s", err)
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %s", err)
	}

	return movies, nil
}

// GetByID returns an uploaded movie with matching ID or an error
func (r *MovieRepository) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	query := `SELECT id, name, picture_key, created_at FROM movies WHERE id = $1`
	var movie models.Movie
	err := r.pool.QueryRow(ctx, query, id).Scan(&movie.ID, &movie.Name, &movie.PictureKey, &movie.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("movie not found")
		}
		return nil, fmt.Errorf("failed to load a movie: %s", err)
	}

	return &movie, nil
}

func (r *MovieRepository) Update(ctx context.Context, id int, name, pictureKey string) error {
	query := `UPDATE movie SET name = $1, picture_key = $2 WHERE id = $3`
	result, err := r.pool.Exec(ctx, query, name, pictureKey, id)
	if err != nil {
		return fmt.Errorf("failed to update movie: %s", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("movie not found")
	}

	return nil
}

func (r *MovieRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM movies WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete movie %d: %s", id, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("movie not found")
	}

	return nil
}
