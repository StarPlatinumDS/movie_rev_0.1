package handlers

import (
	"html/template"
	"log"
	"movie-review/internal/services"
	"net/http"
	"os"
	"strconv"
)

type MovieHandlers struct {
	service *services.MovieService
}

func NewMovieHandlers(service *services.MovieService) *MovieHandlers {
	return &MovieHandlers{service: service}
}

// GetMoviesPage send a page w/ movies and form to the client
func (h *MovieHandlers) GetMoviesPage(w http.ResponseWriter, r *http.Request) {
	movies, err := h.service.GetAllMovies(r.Context())
	if err != nil {
		log.Printf("Error loading movies: %v", err)
		http.Error(w, "Failed to load movies", http.StatusInternalServerError)
		return
	}

	log.Println("Movies loaded:", len(movies))

	// Проверяем текущую директорию
	wd, _ := os.Getwd()
	log.Println("Current working directory:", wd)

	tmpl, err := template.ParseFiles("web/templates/index.gohtml")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, movies)
	if err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// CreateMovie does everything to upload moovie using data from the form
// then redirects
func (h *MovieHandlers) CreateMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit ???
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	file, handler, err := r.FormFile("picture")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	_, err = h.service.CreateMovie(r.Context(), name, file, handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteMovie - deletes movie and redirects
func (h *MovieHandlers) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteMovie(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
