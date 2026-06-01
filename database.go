package main

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// initializing the database
func initDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

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

// save short code against long url
func saveMapping(db *sql.DB, shortCode string, originalURL string) error {
	statement, err := db.Prepare("INSERT INTO urls(short_code, original_url) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(shortCode, originalURL)
	return err
}

func getOriginalURL(db *sql.DB, shortCode string) (string, error) {
	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = ?", shortCode).Scan(&originalURL)
	return originalURL, err
}

func isDuplicateShortCodeError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") &&
		strings.Contains(message, "urls.short_code")
}
