package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestSaveMappingAndGetOriginalURL(t *testing.T) {
	db := setupTestDB(t)

	shortCode := "abc123"
	originalURL := "https://example.com"

	if err := saveMapping(db, shortCode, originalURL); err != nil {
		t.Fatalf("saveMapping returned error: %v", err)
	}

	got, err := getOriginalURL(db, shortCode)
	if err != nil {
		t.Fatalf("getOriginalURL returned error: %v", err)
	}

	if got != originalURL {
		t.Fatalf("expected original URL %q, got %q", originalURL, got)
	}
}

func TestGetOriginalURLNotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := getOriginalURL(db, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestSaveMappingDuplicateShortCode(t *testing.T) {
	db := setupTestDB(t)

	shortCode := "abc123"

	if err := saveMapping(db, shortCode, "https://example.com"); err != nil {
		t.Fatalf("first saveMapping returned error: %v", err)
	}

	err := saveMapping(db, shortCode, "https://example.org")
	if err == nil {
		t.Fatal("expected duplicate short code error, got nil")
	}

	if !isDuplicateShortCodeError(err) {
		t.Fatalf("expected duplicate short code error, got %v", err)
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "urls.db")
	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("initDB returned error: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.Exec("DELETE FROM urls"); err != nil {
			t.Errorf("failed to clean test database state: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	return db
}
