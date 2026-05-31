package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode string `json:"shortCode"`
	ShortURL  string `json:"shortUrl"`
}

func shortenURLHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "shorten handler reached",
	})
}

func redirectURLHandler(c *gin.Context) {
	shortCode := c.Param("shortCode")

	c.JSON(http.StatusOK, gin.H{
		"message":   "redirect handler reached",
		"shortCode": shortCode,
	})
}

func main() {
	fmt.Println("Sample short code:", generateShortCode(6))

	database, err := initDB("urls.db")
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer database.Close()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.POST("/shorten", shortenURLHandler)
	router.GET("/:shortCode", redirectURLHandler)

	router.Run(":8080")
}
