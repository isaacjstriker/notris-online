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
	SupabaseURL string
	SupabaseKey string
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (it's okay if it doesn't exist)
	_ = godotenv.Load()

    config := &Config{
        // Use simple fallbacks, let .env provide the real values
        DatabaseURL:  getEnv("DATABASE_URL", "devware.db"), // Fallback to SQLite
        AppName:      getEnv("APP_NAME", "Dev Ware"),
        Debug:        getEnvBool("DEBUG", false),
        JWTSecret:    getEnv("JWT_SECRET", generateDefaultJWT()),
        ServerPort:   getEnvInt("SERVER_PORT", 8080),
        ServerHost:   getEnv("SERVER_HOST", "localhost"),
        SupabaseURL:  getEnv("SUPABASE_URL", ""),
        SupabaseKey:  getEnv("SUPABASE_KEY", ""),
    }

	return config, nil
}

// Helper function for JWT default
func generateDefaultJWT() string {
	return "YkBSq+6EX9RA4hCTevQFAk1A+YBInBP8eCF8Y5iQP2jPNAuXWXtm4uBaZQYLBHzWW3zjKTwi26PXico13lbcDQ=="
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
