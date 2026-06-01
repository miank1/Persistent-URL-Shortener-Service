package main

import (
	"database/sql"
	"strings"

	_ "modernc.org/sqlite"
)

// initializing the database
func initDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		short_code TEXT PRIMARY KEY,
		original_url TEXT NOT NULL,
		click_count INTEGER DEFAULT 0 NOT NULL
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}

	if err := ensureClickCountColumn(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureClickCountColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(urls)")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}

		if name == "click_count" {
			return nil
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.Exec("ALTER TABLE urls ADD COLUMN click_count INTEGER DEFAULT 0 NOT NULL")
	return err
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

func incrementClickCount(db *sql.DB, shortCode string) error {
	_, err := db.Exec("UPDATE urls SET click_count = click_count + 1 WHERE short_code = ?", shortCode)
	return err
}

func getURLStats(db *sql.DB, shortCode string) (string, int, error) {
	var originalURL string
	var clickCount int
	err := db.QueryRow(
		"SELECT original_url, click_count FROM urls WHERE short_code = ?",
		shortCode,
	).Scan(&originalURL, &clickCount)
	return originalURL, clickCount, err
}

func isDuplicateShortCodeError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") &&
		strings.Contains(message, "urls.short_code")
}
