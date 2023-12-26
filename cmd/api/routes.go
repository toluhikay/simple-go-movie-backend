package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create a router mux
	mux := chi.NewRouter()

	// addigng middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	// adding routes
	mux.Get("/", app.Home)
	mux.Post("/authenticate", app.authenticate)
	mux.Get("/allmovies", app.AllMovies)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logOut)
	mux.Get("/movies/{id}", app.GetOneMovie)
	mux.Get("/allgenres", app.AllGenres)
	mux.Get("/movies/genres/{id}", app.AllMoviesByGenre)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequired)
		mux.Get("/movies", app.MovieCatalogue)
		mux.Get("/movie/{id}", app.GetOneMovieForEdit)
		mux.Put("/movies/0", app.InsertMovie)
		mux.Patch("/movies/{id}", app.UpdateMovie)
		mux.Delete("/movies/{id}", app.DeleteMovie)
	})

	return mux
}
