package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	shortCodeLength      = 6
	shortCodeMaxAttempts = 5
)

// request struct
type ShortenRequest struct {
	URL string `json:"url"`
}

// response struct
type ShortenResponse struct {
	ShortCode string `json:"shortCode"`
	ShortURL  string `json:"shortUrl"`
}

// handler for shortening the url
func shortenURLHandler(db *sql.DB, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ShortenRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		if request.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "url is required",
			})
			return
		}

		if _, err := url.ParseRequestURI(request.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid url",
			})
			return
		}

		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database connection is not initialized",
			})
			return
		}

		var shortCode string

		for attempt := 0; attempt < shortCodeMaxAttempts; attempt++ {
			shortCode = generateShortCode(shortCodeLength)
			if shortCode == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to generate short code",
				})
				return
			}

			if err := saveMapping(db, shortCode, request.URL); err != nil {
				if isDuplicateShortCodeError(err) {
					continue
				}

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to save url mapping",
				})
				return
			}

			response := ShortenResponse{
				ShortCode: shortCode,
				ShortURL:  fmt.Sprintf("%s/%s", baseURL, shortCode),
			}

			c.JSON(http.StatusCreated, response)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate unique short code",
		})
	}
}

// redirect to original url
func redirectURLHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database connection is not initialized",
			})
			return
		}

		shortCode := c.Param("shortCode")
		originalURL, err := getOriginalURL(db, shortCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "short code not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get original url",
			})
			return
		}

		c.Redirect(http.StatusFound, originalURL)
	}
}

func main() {
	fmt.Println("Sample short code:", generateShortCode(6))

	config := loadConfig()

	database, err := initDB(config.DBPath)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer database.Close()

	router := gin.Default()

	// health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// ping
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.POST("/shorten", shortenURLHandler(database, config.BaseURL))
	router.GET("/:shortCode", redirectURLHandler(database))

	router.Run(":" + config.Port)
}
