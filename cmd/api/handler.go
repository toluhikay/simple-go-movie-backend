package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/toluhikay/go-react/internal/models"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Name    string `json:"name"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Name:    "Go React",
		Message: "Destroying golang",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read the json payload
	var reqpayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, reqpayload)
	if err != nil {
		app.errorJSON(w, err)
	}

	// validate the user against the database
	user, err := app.DB.GetUserByEMail(reqpayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credential"))
		return
	}

	// check password
	valid, err := user.PasswordMatch(reqpayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	}

	u := jwtUSer{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate token
	tokens, err := app.auth.GenerateTokens(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// set the cookie and send to the user
	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)

	http.SetCookie(w, refreshCookie)
	app.writeJSON(w, http.StatusOK, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	// first thing is to range over the cookies sent back by the user then check for our own cookie
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parser the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get user id from token claims
			userId, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUSerById(userId)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			// generate a new jwt user
			u := jwtUSer{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			// generate new token pairs
			tokenPairs, err := app.auth.GenerateTokens(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating token"), http.StatusUnauthorized)
				return
			}

			// set a new refresh cookie and send back to user
			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			// write back to use
			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

// create a log out route
func (app *application) logOut(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())

	w.WriteHeader(http.StatusAccepted)
}

func (app *application) MovieCatalogue(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) GetOneMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// convert movie to string
	movieId, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// get the movie from the db
	movie, err := app.DB.GetOneMovie(movieId)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movie)
}

func (app *application) GetOneMovieForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// convert to int as did up there
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// get movie plus genres
	movie, genres, err := app.DB.GetOneMovieForEdit(movieID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// build the payload to return the json
	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		Movie:  movie,
		Genres: genres,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.AllGenres()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, genres)
}

func (app *application) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	err := app.readJSON(w, r, &movie)
	if err != nil {
		app.errorJSON(w, errors.New("invalid payload"))
		return
	}

	// get movie from external resource
	movie = app.getPoster(movie)
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	// insert the movie before handling the payload
	newMovieId, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// handle genres
	err = app.DB.UpdateMovieGenre(newMovieId, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "Movie updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)

}

func (app *application) getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}

	// need a client to make a call to the remote api
	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	//build a new request to teh server

	req, err := http.NewRequest("GET", theUrl+"&query="+url.QueryEscape(movie.Title), nil)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	// add headers to request
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// execute request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	//read response body but defer close the request to avoid resource leak
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	var respBody TheMovieDB

	// unmrashal the body bytes into the response body
	json.Unmarshal(bodyBytes, &respBody)

	// check to see of there are meaningful results from the request
	if len(respBody.Results) > 0 {
		movie.Image = respBody.Results[0].PosterPath
	}
	return movie
}

func (app *application) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var payload models.Movie
	err := app.readJSON(w, r, payload)
	if err != nil {
		app.errorJSON(w, err)
	}

	// get the movie from db with the payload id
	movie, err := app.DB.GetOneMovie(payload.ID)
	if err != nil {
		app.errorJSON(w, err)
	}

	// update the movie gotten from db with the neccessary payload
	movie.Title = payload.Title
	movie.Description = payload.Description
	movie.ReleaseDate = payload.ReleaseDate
	movie.RunTime = payload.RunTime
	movie.MPAARating = payload.MPAARating
	movie.UpdatedAt = payload.UpdatedAt

	// update the movie genre
	err = app.DB.UpdateMovieGenre(movie.ID, payload.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
	}

	// create a response object to return to the user
	resp := JSONResponse{
		Error:   false,
		Message: "movie updated successfully",
	}

	_ = app.writeJSON(w, http.StatusAccepted, resp)

}

func (app *application) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
	}

	err = app.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(w, err)
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie deleted succesfully",
	}

	_ = app.writeJSON(w, http.StatusAccepted, resp)

}

func (app *application) AllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movies, err := app.DB.AllMovies(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movies)
}
