package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"movie-review/internal/models"
	"strings"

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
	query := `INSERT INTO movies (name, picture_key, genres, year, description, rating, world_rating, comment) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	          RETURNING id, created_at`
	err := r.pool.QueryRow(ctx, query,
		movie.Name, movie.PictureKey, movie.Genres, movie.Year,
		movie.Description, movie.Rating, movie.WorldRating, movie.Comment,
	).Scan(&movie.ID, &movie.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert movie: %s", err)
	}
	return nil
}

// GetAll returns a slice of all uploaded movies
func (r *MovieRepository) GetAll(ctx context.Context) ([]models.Movie, error) {
	query := `SELECT id, name, picture_key, genres, year, description, rating, world_rating, comment, created_at 
	          FROM movies ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load movies: %s", err)
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		var worldRating sql.NullFloat64
		err := rows.Scan(&movie.ID, &movie.Name, &movie.PictureKey, &movie.Genres, // нужно чтобы с genres не было беды
			&movie.Year, &movie.Description, &movie.Rating, &worldRating, &movie.Comment, &movie.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to load movie: %s", err)
		}
		if worldRating.Valid {
			movie.WorldRating = float32(worldRating.Float64)
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
	query := `SELECT id, name, picture_key, genres, year, description, rating, world_rating, comment, created_at 
	          FROM movies WHERE id = $1`
	var movie models.Movie
	var worldRating sql.NullFloat64
	err := r.pool.QueryRow(ctx, query, id).Scan(&movie.ID, &movie.Name, &movie.PictureKey, &movie.Genres,
		&movie.Year, &movie.Description, &movie.Rating, &worldRating, &movie.Comment, &movie.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("movie not found")
		}
		return nil, fmt.Errorf("failed to load a movie: %s", err)
	}
	if worldRating.Valid {
		movie.WorldRating = float32(worldRating.Float64)
	}

	return &movie, nil
}

// SearchMovies - search on /search page with conditions and pagination
func (r *MovieRepository) SearchMovies(ctx context.Context, query string, yearFrom int, yearTo int, genres []string, hasComment *bool, offset, limit int) ([]models.Movie, error) {
	querySQL := `SELECT id, name, picture_key, genres, year, description, rating, world_rating, comment, created_at FROM movies`

	var whereClauses []string
	var args []interface{}
	argCount := 1

	// name filter
	if query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name LIKE $%d", argCount))
		args = append(args, "%"+query+"%")
		argCount++
	}

	// year FROM filter
	if yearFrom > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("year >= $%d", argCount))
		args = append(args, yearFrom)
		argCount++
	}

	// year TO filter
	if yearTo > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("year <= $%d", argCount))
		args = append(args, yearTo)
		argCount++
	}

	// genres filter
	if len(genres) > 0 {
		// there may be an array overlap therefore I use &&
		placeholders := make([]string, len(genres))
		for i, genre := range genres {
			placeholders[i] = fmt.Sprintf("$%d", argCount+i)
			args = append(args, genre)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("genres && ARRAY[%s]", strings.Join(placeholders, ",")))
		argCount += len(genres)
	}

	// hasComment filter
	if hasComment != nil {
		if *hasComment {
			whereClauses = append(whereClauses, fmt.Sprintf("comment IS NOT NULL AND comment != ''"))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("comment IS NULL OR comment = ''"))
		}
	}

	// that's query's WHERE part
	if len(whereClauses) > 0 {
		querySQL += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sorting: if query exists we search by relevance, otherwise - by date
	if query != "" {
		querySQL += " ORDER BY CASE WHEN name LIKE $1 THEN 0 ELSE 1 END, created_at DESC"
	} else {
		querySQL += " ORDER BY created_at DESC"
	}

	// pagination
	querySQL += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	// make a request
	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %s", err)
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		var worldRating sql.NullFloat64
		err := rows.Scan(&movie.ID, &movie.Name, &movie.PictureKey, &movie.Genres,
			&movie.Year, &movie.Description, &movie.Rating, &worldRating, &movie.Comment, &movie.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %s", err)
		}
		if worldRating.Valid {
			movie.WorldRating = float32(worldRating.Float64)
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %s", err)
	}

	return movies, nil
}

func (r *MovieRepository) Update(ctx context.Context, movie *models.Movie) error {
	query := `UPDATE movies SET name = $1, picture_key = $2, genres = $3, year = $4, 
	          description = $5, rating = $6, world_rating = $7, comment = $8 
	          WHERE id = $9`
	result, err := r.pool.Exec(ctx, query, movie.Name, movie.PictureKey, movie.Genres, movie.Year,
		movie.Description, movie.Rating, movie.WorldRating, movie.Comment, movie.ID)
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
