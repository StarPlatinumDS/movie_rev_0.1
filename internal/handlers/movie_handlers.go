package handlers

import (
	"html/template"
	"log"
	"mime/multipart"
	"movie-review/internal/models"
	"movie-review/internal/services"
	"net/http"
	"strconv"
)

// helper functions for templates
var funcMap = template.FuncMap{
	"iterate": func(count int) []int {
		var items []int
		for i := 0; i < count; i++ {
			items = append(items, i)
		}
		return items
	},
	"inSlice": func(item string, list []string) bool {
		for _, v := range list {
			if v == item {
				return true
			}
		}
		return false
	},
	"sub": func(a, b int) int {
		return a - b
	},
}

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

	tmpl := template.New("index.gohtml").Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles("web/templates/index.gohtml")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Movies []models.MovieResponse
		Genres []string
	}{
		Movies: movies,
		Genres: models.AvailableGenres,
	}

	err = tmpl.Execute(w, data)
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
	year, _ := strconv.Atoi(r.FormValue("year"))
	description := r.FormValue("description")
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	worldRating, _ := strconv.ParseFloat(r.FormValue("world_rating"), 32)
	comment := r.FormValue("comment")
	genres := r.Form["genres"]

	file, handler, err := r.FormFile("picture")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	_, err = h.service.CreateMovie(r.Context(), name, file, handler.Filename, genres, year, description,
		rating, float32(worldRating), comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GetMovieDetail sends to the client a detailed
// single movie page with all the info
func (h *MovieHandlers) GetMovieDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	movie, err := h.service.GetMovieByID(r.Context(), id)
	if err != nil {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}

	tmpl := template.New("movie_detail.gohtml").Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles("web/templates/movie_detail.gohtml")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Movie  *models.MovieResponse
		Genres []string
	}{
		Movie:  movie,
		Genres: models.AvailableGenres,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// GetMovieEdit - creates and sends back to the user edit form
func (h *MovieHandlers) GetMovieEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	movie, err := h.service.GetMovieByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	tmpl := template.New("movie_edit.gohtml").Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles("web/templates/movie_edit.gohtml")
	if err != nil {
		log.Printf("Tempate parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Movie  *models.MovieResponse
		Genres []string
	}{
		Movie:  movie,
		Genres: models.AvailableGenres,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Teamplate execute error: %v", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

func (h *MovieHandlers) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	year, _ := strconv.Atoi(r.FormValue("year"))
	description := r.FormValue("description")
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	worldRating, _ := strconv.ParseFloat(r.FormValue("world_rating"), 32)
	comment := r.FormValue("comment")
	genres := r.Form["genres"]

	var file multipart.File
	var handler *multipart.FileHeader
	file, handler, err = r.FormFile("picture")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}

	var updatedMovie *models.MovieResponse
	if file != nil && handler != nil {
		defer file.Close()
		updatedMovie, err = h.service.UpdateMovie(r.Context(), id, name, file, handler.Filename, genres, year,
			description, rating, float32(worldRating), comment)
	} else {
		// update without picture swap
		updatedMovie, err = h.service.UpdateMovie(r.Context(), id, name, nil, "", genres, year,
			description, rating, float32(worldRating), comment)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/movie/"+strconv.Itoa(updatedMovie.ID), http.StatusSeeOther)
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
