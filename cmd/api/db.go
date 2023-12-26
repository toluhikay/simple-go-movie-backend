package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func openDb(dsn string) (*sql.DB, error) {
	// open the database
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *application) connectToDb() (*sql.DB, error) {
	connection, err := openDb(app.DSN)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to postgres successfully")
	return connection, nil
}
