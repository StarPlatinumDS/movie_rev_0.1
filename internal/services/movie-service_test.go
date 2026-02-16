package services

import (
	"context"
	"errors"
	"testing"

	"movie-review/internal/mocks"
	"movie-review/internal/models"

	"github.com/stretchr/testify/assert"
)

// Unit tests for the service
func TestMovieService_GetMovieByID_Success(t *testing.T) {

	// create mocks
	mockRepo := new(mocks.MovieRepository)
	mockMinio := new(mocks.FileStorage)

	// set behavior
	ctx := context.Background()
	expectedMovie := &models.Movie{
		ID:         1,
		Name:       "Test Movie",
		PictureKey: "img/1.jpg",
	}

	// Wait for .GetByID() and return daa
	mockRepo.On("GetByID", ctx, 1).Return(expectedMovie, nil)

	// init mock service
	service := NewMovieService(mockRepo, mockMinio, "bucket", "http://minio")

	// call the method
	result, err := service.GetMovieByID(ctx, 1)

	// check the result
	assert.NoError(t, err)
	assert.Equal(t, "Test Movie", result.Name)

	// checl that all calls have been made
	mockRepo.AssertExpectations(t)
}

func TestMovieService_GetMovieByID_NotFound(t *testing.T) {

	// create mocks
	mockRepo := new(mocks.MovieRepository)
	mockMinio := new(mocks.FileStorage)

	// set behaviour
	ctx := context.Background()
	// simulate the db err
	mockRepo.On("GetByID", ctx, 999).Return(nil, errors.New("not found"))

	// init mock service
	service := NewMovieService(mockRepo, mockMinio, "bucket", "http://minio")

	// call the method
	result, err := service.GetMovieByID(ctx, 999)

	// check the result
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
