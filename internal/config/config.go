package config

import (
	"os"
	"strconv"
	"fmt"
	"crypto/rand"
	"encoding/base64"

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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateSecureJWT()
		fmt.Println("🔐 Generated new JWT secret. Add this to your .env file:")
		fmt.Printf("JWT_SECRET=%s\n", jwtSecret)
	}

    config := &Config{
        // Use simple fallbacks, let .env provide the real values
        DatabaseURL:  getEnv("DATABASE_URL", "devware.db"), // Fallback to SQLite
        AppName:      getEnv("APP_NAME", "Dev Ware"),
        Debug:        getEnvBool("DEBUG", false),
        JWTSecret:    jwtSecret,
        ServerPort:   getEnvInt("SERVER_PORT", 8080),
        ServerHost:   getEnv("SERVER_HOST", "localhost"),
        SupabaseURL:  getEnv("SUPABASE_URL", ""),
        SupabaseKey:  getEnv("SUPABASE_KEY", ""),
    }

	return config, nil
}

//generateSecureJWT creates a cryptographically secure JWT secret
func generateSecureJWT() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		fmt.Println("⚠️  CRITICAL: Could not generate secure JWT secret!")
        os.Exit(1)
    }
    return base64.StdEncoding.EncodeToString(bytes)
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
