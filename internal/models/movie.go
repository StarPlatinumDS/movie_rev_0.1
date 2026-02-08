package models

import "time"

// Main struct to interract with db
type Movie struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	PictureKey string    `json:"pictureKey"`
	CreatedAt  time.Time `json:"createdAt"`
}

// What the user will send to us through form
// on the website
type MovieCreateRequest struct {
	Name    string `json:"name" form:"name"`
	Picture []byte `json:"-" form:"picture"`
}

// What we'll send back to the user
type MovieResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PictureURL string `json:"pictureURL"`
}
