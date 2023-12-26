package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/toluhikay/go-react/internal/models"
)

// this struct hold all of the database connections
type PostgresDbRepo struct {
	DB *sql.DB
}

const dbTimeOut = time.Second * 10

// create a function that can close connection
func (m *PostgresDbRepo) Connection() *sql.DB {
	return m.DB
}

// create a function that will make it implement the database repo
func (m *PostgresDbRepo) AllMovies(genre ...int) ([]*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// create a where clause to handle the if there is any genre supplied
	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from movies_genres where genre_id = %d)", genre[0])
	}

	// sql query to interact with db
	query := fmt.Sprintf(`
		select
			id, title, release_date, runtime,
			mpaa_rating, description, coalesce(image, ''),
			created_at, updated_at
		from
			movies %s
		order by
			title

	`, where)

	// query the db now for the rows
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// close row to avoid db leak
	defer rows.Close()

	var movies []*models.Movie

	// iterate of the rows and scan the rows into the movie slice
	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.RunTime,
			&movie.MPAARating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)

	}

	return movies, nil
}

func (m *PostgresDbRepo) GetOneMovie(id int) (*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	query := `select id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		from movies where id = $1
	`
	var movie models.Movie
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// get genre if any
	query = `select g.id, g.genre from movies_genres mg
			left join genres g on (mg.genre_id = g.id)
			where mg.movie_id = $1
			order by g.genre
	`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// close rows to avoid resource leak
	defer rows.Close()

	var genres []*models.Genre
	// loop through rows to store each value
	for rows.Next() {
		// variable to store individul genre
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)

		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}
	movie.Genres = genres
	return &movie, nil

}

func (m *PostgresDbRepo) GetOneMovieForEdit(id int) (*models.Movie, []*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	query := `select id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		from movies where id = $1
	`
	var movie models.Movie
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	// get genre if any
	query = `select g.id, g.genre from movies_genres mg
			left join genres g on (mg.genre_id = g.id)
			where mg.movie_id = $1
			order by g.genre
	`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	// close rows to avoid resource leak
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int
	// loop through rows to store each value
	for rows.Next() {
		// variable to store individul genre
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)

		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.ID)
	}
	movie.Genres = genres
	movie.GenresArray = genresArray

	var allGenres []*models.Genre
	query = `select id, genre from genres order by genre`

	gRows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer gRows.Close()

	for gRows.Next() {
		var g models.Genre
		err := gRows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}
		allGenres = append(allGenres, &g)
	}

	return &movie, allGenres, nil

}

func (m *PostgresDbRepo) GetUserByEMail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// create the query
	query := `select id, email, first_name, last_name, password, created_at, updated_at
			from users where email = $1
	`
	// scan the user into a row
	var user models.User
	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.Email,
		&user.Password,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, err
}

func (m *PostgresDbRepo) GetUSerById(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// query db with user id
	query := `select id, email, first_name, last_name, password, created_at, updated_at
					from users where id = $1`

	var user models.User
	// get the row of the user if available
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.Email,
		&user.Password,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *PostgresDbRepo) AllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	query := `select id, genre, created_at, updated_at from genres order by genre`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var genres []*models.Genre

	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Genre,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		genres = append(genres, &genre)
	}
	return genres, nil
}

func (m *PostgresDbRepo) InsertMovie(movie models.Movie) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	var newMovieID int

	stmt := `insert into movies (id, title, description, release_date, runtime, mpaa_rating, created_at, updated_at, image)
			values ($1,$2, $3,$4, $5, $6, $7, $8) returning id
	`

	err := m.DB.QueryRowContext(ctx, stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedAt,
		movie.Image,
	).Scan(&newMovieID)

	if err != nil {
		return 0, err
	}

	return newMovieID, nil
}

func (m *PostgresDbRepo) UpdateMovie(movie models.Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	stmt := `update movies set title = $1, description = $2, release_date = $3, 
				runtime = $4, mpaa_rating = $5, 
				updated_at = $6, image = $7 where id = $8`

	_, err := m.DB.ExecContext(ctx, stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
		movie.MPAARating,
		movie.UpdatedAt,
		movie.Image,
		movie.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresDbRepo) UpdateMovieGenre(id int, genreIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// for this purpose first delete the movie genres id
	stmt := `delete from movie_genres where movie_d = $1`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	// range through the genre Ids and insert into movie genres
	for _, n := range genreIDs {
		stmt := `insert into movie_genres (movie_id, genre_id) values ($1, $2)`
		_, err := m.DB.ExecContext(ctx, stmt, id, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *PostgresDbRepo) DeleteMovie(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	stmt := `delete from movies where id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
