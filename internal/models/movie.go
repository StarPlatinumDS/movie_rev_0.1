package models

import "time"

// All avaible genres here
var AvailableGenres = []string{
	"Боевик", "Приключения", "Мультфильм", "Биография", "Комедия", "Криминал",
	"Документальный", "Драма", "Семейный", "Фэнтези", "Исторический", "Ужасы",
	"Музыка", "Мюзикл", "Детектив", "Мелодрама", "Научная фантастика", "Спорт",
	"Триллер", "Военный", "Вестерн",
}

// Main struct to interract with db
type Movie struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PictureKey  string    `json:"picture_key"`
	Genres      []string  `json:"genres"`
	Year        int       `json:"year"`
	Description string    `json:"description"`
	Rating      int       `json:"rating"`
	WorldRating float32   `json:"world_rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

// What the user will send to us through form
// on the website
type MovieCreateRequest struct {
	Name    string `json:"name" form:"name"`
	Picture []byte `json:"-" form:"picture"`
}

// What we'll send back to the user (includes picture)
type MovieResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	PictureURL  string   `json:"picture_url"`
	Genres      []string `json:"genres"`
	Year        int      `json:"year"`
	Description string   `json:"description"`
	Rating      int      `json:"rating"`
	WorldRating float32  `json:"world_rating"`
	Comment     string   `json:"comment"`
	HasComment  bool     `json:"has_comment"`
}
