package main

import "os"

const (
	defaultPort   = "8080"
	defaultDBPath = "urls.db"
)

type Config struct {
	Port    string
	DBPath  string
	BaseURL string
}

func loadConfig() Config {
	port := getEnv("PORT", defaultPort)

	return Config{
		Port:    port,
		DBPath:  getEnv("DB_PATH", defaultDBPath),
		BaseURL: getEnv("BASE_URL", "http://localhost:"+port),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
