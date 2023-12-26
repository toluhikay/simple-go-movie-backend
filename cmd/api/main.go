package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/toluhikay/go-react/internal/repository"
	dbrepo "github.com/toluhikay/go-react/internal/repository/dbRepo"
)

type application struct {
	DSN          string
	Domain       string
	DB           repository.DatabaseRepo
	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
	APIKey       string
}

func main() {
	var app application

	// read from comand line using the flag package, second arg is what i want the flag to be on the cmd line
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5433 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC+1 connect_timeout=5", "Postgres connection string")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "example.com", "signing secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", "example.com", "signing issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", "example.com", "signing audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie domain")
	flag.StringVar(&app.Domain, "domain", "example.com", "domain")
	flag.StringVar(&app.APIKey, "api-key", "4afe5bb347ffcc8555b9646caac7b88d", "api key")
	flag.Parse()

	// connect to db using pgx v4
	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}

	app.DB = &dbrepo.PostgresDbRepo{DB: conn}

	// defer conn.Close() -> one way to close conn another is down
	defer app.DB.Connection().Close()

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "__Host-refresh_token",
		CookieDomain:  app.CookieDomain,
	}

	fmt.Println("Listening on port 4000")
	log.Fatal(http.ListenAndServe(":"+"4000", app.routes()))
}
