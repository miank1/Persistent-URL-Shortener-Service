package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

type StatsResponse struct {
	OriginalURL string `json:"originalUrl"`
	ClickCount  int    `json:"clickCount"`
}

// handler for shortening the url
func shortenURLHandler(db *sql.DB, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logRequest(c, "shorten_url")

		var request ShortenRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			log.Warn().
				Err(err).
				Int("status", http.StatusBadRequest).
				Msg("shorten request failed validation")
			respondWithError(c, http.StatusBadRequest, "invalid request body")
			return
		}

		if request.URL == "" {
			log.Warn().
				Int("status", http.StatusBadRequest).
				Msg("shorten request missing url")
			respondWithError(c, http.StatusBadRequest, "url is required")
			return
		}

		if _, err := url.ParseRequestURI(request.URL); err != nil {
			log.Warn().
				Err(err).
				Int("status", http.StatusBadRequest).
				Msg("shorten request has invalid url")
			respondWithError(c, http.StatusBadRequest, "invalid url")
			return
		}

		if db == nil {
			log.Error().
				Int("status", http.StatusInternalServerError).
				Msg("shorten request failed because database is not initialized")
			respondWithError(c, http.StatusInternalServerError, "database connection is not initialized")
			return
		}

		var shortCode string

		for attempt := 0; attempt < shortCodeMaxAttempts; attempt++ {
			shortCode = generateShortCode(shortCodeLength)
			if shortCode == "" {
				log.Error().
					Int("status", http.StatusInternalServerError).
					Msg("shorten request failed to generate short code")
				respondWithError(c, http.StatusInternalServerError, "failed to generate short code")
				return
			}

			if err := saveMapping(db, shortCode, request.URL); err != nil {
				if isDuplicateShortCodeError(err) {
					log.Warn().
						Err(err).
						Str("short_code", shortCode).
						Int("attempt", attempt+1).
						Msg("generated duplicate short code")
					continue
				}

				log.Error().
					Err(err).
					Int("status", http.StatusInternalServerError).
					Msg("shorten request failed to save url mapping")
				respondWithError(c, http.StatusInternalServerError, "failed to save url mapping")
				return
			}

			response := ShortenResponse{
				ShortCode: shortCode,
				ShortURL:  baseURL + "/" + shortCode,
			}

			log.Info().
				Str("short_code", response.ShortCode).
				Str("short_url", response.ShortURL).
				Int("status", http.StatusCreated).
				Msg("shorten request completed")
			c.JSON(http.StatusCreated, response)
			return
		}

		log.Error().
			Int("status", http.StatusInternalServerError).
			Int("attempts", shortCodeMaxAttempts).
			Msg("shorten request failed to generate unique short code")
		respondWithError(c, http.StatusInternalServerError, "failed to generate unique short code")
	}
}

// redirect to original url
func redirectURLHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logRequest(c, "redirect_url")

		shortCode := c.Param("shortCode")

		if db == nil {
			log.Error().
				Str("short_code", shortCode).
				Int("status", http.StatusInternalServerError).
				Msg("redirect request failed because database is not initialized")
			respondWithError(c, http.StatusInternalServerError, "database connection is not initialized")
			return
		}

		originalURL, err := getOriginalURL(db, shortCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn().
					Str("short_code", shortCode).
					Int("status", http.StatusNotFound).
					Msg("redirect request short code not found")
				respondWithError(c, http.StatusNotFound, "short code not found")
				return
			}

			log.Error().
				Err(err).
				Str("short_code", shortCode).
				Int("status", http.StatusInternalServerError).
				Msg("redirect request failed to get original url")
			respondWithError(c, http.StatusInternalServerError, "failed to get original url")
			return
		}

		if err := incrementClickCount(db, shortCode); err != nil {
			log.Error().
				Err(err).
				Str("short_code", shortCode).
				Msg("redirect request failed to increment click count")
		}

		log.Info().
			Str("short_code", shortCode).
			Str("original_url", originalURL).
			Int("status", http.StatusFound).
			Msg("redirect request completed")
		c.Redirect(http.StatusFound, originalURL)
	}
}

func getStatsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logRequest(c, "get_stats")

		shortCode := c.Param("shortCode")

		if db == nil {
			log.Error().
				Str("short_code", shortCode).
				Int("status", http.StatusInternalServerError).
				Msg("stats request failed because database is not initialized")
			respondWithError(c, http.StatusInternalServerError, "database connection is not initialized")
			return
		}

		originalURL, clickCount, err := getURLStats(db, shortCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn().
					Str("short_code", shortCode).
					Int("status", http.StatusNotFound).
					Msg("stats request short code not found")
				respondWithError(c, http.StatusNotFound, "short code not found")
				return
			}

			log.Error().
				Err(err).
				Str("short_code", shortCode).
				Int("status", http.StatusInternalServerError).
				Msg("stats request failed to get url stats")
			respondWithError(c, http.StatusInternalServerError, "failed to get url stats")
			return
		}

		response := StatsResponse{
			OriginalURL: originalURL,
			ClickCount:  clickCount,
		}

		log.Info().
			Str("short_code", shortCode).
			Int("click_count", clickCount).
			Int("status", http.StatusOK).
			Msg("stats request completed")
		c.JSON(http.StatusOK, response)
	}
}

func respondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": message,
	})
}

func logRequest(c *gin.Context, handler string) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	log.Info().
		Str("handler", handler).
		Str("method", c.Request.Method).
		Str("path", path).
		Str("client_ip", c.ClientIP()).
		Msg("received request")
}

func configureLogger() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(os.Stdout).With().Timestamp().Logger()
}

func main() {
	configureLogger()

	log.Info().Str("short_code", generateShortCode(6)).Msg("generated sample short code")

	config := loadConfig()

	database, err := initDB(config.DBPath)
	if err != nil {
		log.Fatal().Err(err).Str("db_path", config.DBPath).Msg("failed to initialize database")
	}
	defer database.Close()

	router := gin.Default()

	// health check
	router.GET("/health", func(c *gin.Context) {
		logRequest(c, "health")
		log.Info().
			Int("status", http.StatusOK).
			Msg("health check completed")
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// ping
	router.GET("/ping", func(c *gin.Context) {
		logRequest(c, "ping")
		log.Info().
			Int("status", http.StatusOK).
			Msg("ping completed")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.POST("/shorten", shortenURLHandler(database, config.BaseURL))
	router.GET("/stats/:shortCode", getStatsHandler(database))
	router.GET("/:shortCode", redirectURLHandler(database))

	if err := router.Run(":" + config.Port); err != nil {
		log.Fatal().Err(err).Str("port", config.Port).Msg("failed to start server")
	}
}
