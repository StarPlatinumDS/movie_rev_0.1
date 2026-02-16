package database

import (
	"context"
	"os"
	"testing"
	"time"

	"movie-review/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontPg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// integration tests. Testing SQL requests using docker containers

// setupTestDB should return conn string
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {

	ctx := context.Background()

	// launch pg container
	postgresContainer, err := testcontPg.Run(
		ctx,
		"postgres:15-alpine",
		testcontPg.WithDatabase("testdb"),
		testcontPg.WithUsername("test"),
		testcontPg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// do migrations
	migrationSQL, err := os.ReadFile("../../init-scripts/01-init.sql")
	require.NoError(t, err)

	// exec SQL
	_, err = pool.Exec(ctx, string(migrationSQL))
	require.NoError(t, err)

	// cleanup func
	cleanup := func() {
		pool.Close()
		postgresContainer.Terminate(ctx)
	}

	return pool, cleanup
}

func TestMovieRepository_Create(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	// run migrations
	repo := NewMovieRepository(pool)
	ctx := context.Background()

	movie := &models.Movie{
		Name:        "Inception",
		PictureKey:  "img/inception.jpg", // Обязательно: NOT NULL
		Year:        2010,
		Description: "A thief who steals corporate secrets through dream-sharing technology",
		Rating:      5,   // Обязательно: CHECK (1-5)
		WorldRating: 8.8, // Обязательно: CHECK (0.0-10.0)
		Comment:     "Great movie!",
		Genres:      []string{"Sci-Fi", "Action"},
	}
	err := repo.Create(ctx, movie)

	require.NoError(t, err)
	assert.Greater(t, movie.ID, 0) //check ID's assigned
}
