package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"movie-review/internal/models"
	"movie-review/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

type SearchHandler struct {
	service repository.MovieService
}

func NewSearchHandler(service repository.MovieService) *SearchHandler {
	return &SearchHandler{
		service: service,
	}
}

// SearchPage renders and sends back to the client
// search page with filters
func (h *SearchHandler) SearchPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	genres, err := h.service.GetAllGenres(ctx)
	if err != nil {
		log.Printf("Error loading genres: %v", err)
		http.Error(w, "Failed to load genres", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/search.gohtml")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Genres []string
	}{
		Title:  "Поиск фильмов",
		Genres: genres,
	}

	err = tmpl.Execute(w, data) // ← ИСПРАВЛЕНО: Execute вместо ExecuteTemplate
	if err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
	}
}

// SearchAPI - live search JSON API
func (h *SearchHandler) SearchAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// parse params
	query := r.URL.Query().Get("query")

	yearFrom, _ := strconv.Atoi(r.URL.Query().Get("yearFrom"))
	yearTo, _ := strconv.Atoi(r.URL.Query().Get("yearTo"))

	genreStr := r.URL.Query().Get("genres")
	var genres []string
	if genreStr != "" {
		genres = strings.Split(genreStr, ",")
	}

	hasCommentStr := r.URL.Query().Get("hasComment")
	var hasComment *bool
	// wtf??
	if hasCommentStr == "true" {
		// true - only with comments
		yes := true
		hasComment = &yes
	} else if hasCommentStr == "false" {
		no := false
		hasComment = &no
	}

	// pagination
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20 // default
	}

	// run search
	movies, err := h.service.SearchMovies(ctx, query, yearFrom, yearTo, genres, hasComment, offset, limit)
	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// form a response
	response := struct {
		Result  []models.MovieResponse `json:"results"`
		HasMore bool                   `json:"hasMore"`
	}{
		Result:  movies,
		HasMore: len(movies) >= limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "JSON encode error", http.StatusInternalServerError)
	}
}
