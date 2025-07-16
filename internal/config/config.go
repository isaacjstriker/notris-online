package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppName     string
	Debug       bool
	JWTSecret   string
	ServerPort  int
	ServerHost  string
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (it's okay if it doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "sqlite3://./devware.db"),
		AppName:     getEnv("APP_NAME", "Dev Ware"),
		Debug:       getEnvBool("DEBUG", true),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		ServerPort:  getEnvInt("SERVER_PORT", 8080),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
	}

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
