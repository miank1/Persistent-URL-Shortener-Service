package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

const testBaseURL = "http://localhost:8080"

func TestShortenURLHandlerCreatesShortURL(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	response := performRequest(router, http.MethodPost, "/shorten", []byte(`{"url":"https://example.com"}`))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var body ShortenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body.ShortCode) != shortCodeLength {
		t.Fatalf("expected short code length %d, got %d", shortCodeLength, len(body.ShortCode))
	}

	expectedShortURL := testBaseURL + "/" + body.ShortCode
	if body.ShortURL != expectedShortURL {
		t.Fatalf("expected short URL %q, got %q", expectedShortURL, body.ShortURL)
	}

	originalURL, err := getOriginalURL(db, body.ShortCode)
	if err != nil {
		t.Fatalf("expected saved URL mapping, got error: %v", err)
	}

	if originalURL != "https://example.com" {
		t.Fatalf("expected original URL %q, got %q", "https://example.com", originalURL)
	}
}

func TestShortenURLHandlerRejectsInvalidRequestBody(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	response := performRequest(router, http.MethodPost, "/shorten", []byte(`{`))

	assertErrorResponse(t, response, http.StatusBadRequest, "invalid request body")
}

func TestRedirectURLHandlerRedirectsToOriginalURL(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	shortCode := "abc123"
	originalURL := "https://example.com"

	if err := saveMapping(db, shortCode, originalURL); err != nil {
		t.Fatalf("saveMapping returned error: %v", err)
	}

	response := performRequest(router, http.MethodGet, "/"+shortCode, nil)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}

	if location := response.Header().Get("Location"); location != originalURL {
		t.Fatalf("expected Location header %q, got %q", originalURL, location)
	}
}

func TestRedirectURLHandlerReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	response := performRequest(router, http.MethodGet, "/missing", nil)

	assertErrorResponse(t, response, http.StatusNotFound, "short code not found")
}

func setupTestRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/shorten", shortenURLHandler(db, testBaseURL))
	router.GET("/:shortCode", redirectURLHandler(db))

	return router
}

func performRequest(router http.Handler, method string, path string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	t.Helper()

	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, response.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["error"] != expectedMessage {
		t.Fatalf("expected error %q, got %q", expectedMessage, body["error"])
	}
}
