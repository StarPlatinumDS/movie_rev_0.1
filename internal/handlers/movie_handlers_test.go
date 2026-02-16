package handlers

import (
	"movie-review/internal/mocks"
	"movie-review/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// test api
func TestGetMovieDetailHandler(t *testing.T) {
	// service mock
	mockService := new(mocks.MovieService)

	// expectations
	mockService.On("GetMovieByID", mock.Anything, 1).Return(&models.MovieResponse{
		ID:   1,
		Name: "Test Movie",
	}, nil)

	// create a handler
	handler := NewMovieHandlers(mockService)

	// create mux
	mux := http.NewServeMux()
	mux.HandleFunc("/movie/{id}", handler.GetMovieDetail)

	// test req
	req := httptest.NewRequest(http.MethodGet, "/movie/1", nil)
	w := httptest.NewRecorder()

	// call handler method
	mux.ServeHTTP(w, req)

	// check mock call
	mockService.AssertExpectations(t)

	assert.NotEqual(t, http.StatusBadRequest, w.Code, "Handler should parse ID correctly")
}
