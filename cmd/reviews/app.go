package main

import (
	"context"
	"fmt"
	"log"
	"movie-review/internal/database"
	"movie-review/internal/handlers"
	"movie-review/internal/services"
	"movie-review/internal/storage"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type App struct {
	httpServer *http.Server
	dbPool     *pgxpool.Pool
}

func NewApp() *App {
	// Загружаем .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// === PostgreSQL ===
	dbPool, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// УБРАЛ: defer dbPool.Close() — пул должен жить всё время

	// === MinIO ===
	minioClient, err := initMinio()
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	// === Storage ===
	minioStorage := storage.NewMinioStorage(minioClient)

	// === Repository ===
	movieRepo := database.NewMovieRepository(dbPool)

	// === Service ===
	bucket := os.Getenv("MINIO_DEFAULT_BUCKET")
	if bucket == "" {
		bucket = "my-bucket"
	}

	publicURL := os.Getenv("MINIO_PUBLIC_URL")
	if publicURL == "" {
		publicURL = fmt.Sprintf("http://localhost:%s/%s", os.Getenv("MINIO_API_PORT"), bucket)
	}

	movieService := services.NewMovieService(movieRepo, minioStorage, bucket, publicURL)

	// === Handlers ===
	movieHandlers := handlers.NewMovieHandlers(movieService)

	// === Router ===
	mux := http.NewServeMux()

	// Статика
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Роуты
	mux.HandleFunc("/", movieHandlers.GetMoviesPage)
	mux.HandleFunc("POST /movies", movieHandlers.CreateMovie)
	mux.HandleFunc("GET /movie/{id}", movieHandlers.GetMovieDetail)
	mux.HandleFunc("GET /movie/{id}/edit", movieHandlers.GetMovieEdit)
	mux.HandleFunc("POST /movie/{id}/edit", movieHandlers.UpdateMovie)
	mux.HandleFunc("GET /movies/delete", movieHandlers.DeleteMovie)

	// === Server ===
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &App{
		httpServer: server,
		dbPool:     dbPool,
	}
}

func (a *App) Run() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(a.httpServer.ListenAndServe())
}

func (a *App) Close() {
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}

// initDatabase создаёт пул соединений с PostgreSQL
func initDatabase() (*pgxpool.Pool, error) {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("✓ PostgreSQL connected")
	return pool, nil
}

// initMinio создаёт клиент для MinIO
func initMinio() (*minio.Client, error) {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ROOT_USER")
	minioSecretKey := os.Getenv("MINIO_ROOT_PASSWORD")

	useSSL := false

	client, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create MinIO client: %w", err)
	}

	// Проверяем соединение
	_, err = client.ListBuckets(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to ping MinIO: %w", err)
	}

	log.Println("✓ MinIO connected")
	return client, nil
}
