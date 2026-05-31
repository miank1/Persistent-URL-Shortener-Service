package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB(filepath string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	db = database

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		short_code TEXT PRIMARY KEY,
		original_url TEXT NOT NULL
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}

	return db, nil
}
