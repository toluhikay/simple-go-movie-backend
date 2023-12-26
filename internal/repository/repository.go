package repository

import (
	"database/sql"

	"github.com/toluhikay/go-react/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	AllMovies(genre ...int) ([]*models.Movie, error)
	GetUserByEMail(email string) (*models.User, error)
	GetUSerById(id int) (*models.User, error)

	GetOneMovie(id int) (*models.Movie, error)
	GetOneMovieForEdit(id int) (*models.Movie, []*models.Genre, error)
	AllGenres() ([]*models.Genre, error)
	InsertMovie(movie models.Movie) (int, error)
	UpdateMovieGenre(id int, genreIDs []int) error
	UpdateMovie(movie models.Movie) error
	DeleteMovie(id int) error
}
