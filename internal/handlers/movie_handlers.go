package handlers

import (
	"html/template"
	"movie-review/internal/services"
	"net/http"

	"github.com/google/uuid"
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
		http.Error(w, "Failed to load movies", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("web/templates/index.gohtml"))
	tmpl.Execute(w, movies)
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteMovie(r.Context(), int(id.ID()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
