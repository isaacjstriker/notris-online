package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
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

// Load reads configuration from environment variables or a .env file
func Load() (*Config, error) {
	// Load .env file if it exists (useful for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, reading from environment")
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppName:     getEnv("APP_NAME", "Dev Ware"),
		Debug:       getEnvAsBool("DEBUG", false),
		ServerPort:  getEnvAsInt("SERVER_PORT", 8080),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
		JWTSecret:   os.Getenv("JWT_SECRET"), // Load the secret
	}

	// --- VALIDATION AND AUTO-CONFIGURATION LOGIC ---
	if cfg.JWTSecret == "" {
		// Generate a new key.
		newKey := make([]byte, 32) // 256 bits
		if _, err := rand.Read(newKey); err != nil {
			return nil, fmt.Errorf("failed to generate a new JWT key: %w", err)
		}
		encodedKey := base64.StdEncoding.EncodeToString(newKey)

		// Attempt to create/append to the .env file automatically.
		envFilePath := ".env"
		f, err := os.OpenFile(envFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// If we can't write to the file (e.g., permissions error), fall back to manual instructions.
			errorMsg := fmt.Sprintf(`
FATAL: JWT_SECRET is not set and I couldn't write to the .env file.
Error: %v

Please create a .env file and add the following line:

JWT_SECRET=%s

`, err, encodedKey)
			return nil, fmt.Errorf("%s", errorMsg)
		}
		defer f.Close()

		// Write the new secret to the file.
		newLine := fmt.Sprintf("\nJWT_SECRET=%s\n", encodedKey)
		if _, err := f.WriteString(newLine); err != nil {
			// If writing fails, fall back to manual instructions.
			errorMsg := fmt.Sprintf(`
FATAL: JWT_SECRET is not set and I failed to write to the .env file.
Error: %v

Please add the following line to your .env file:

JWT_SECRET=%s

`, err, encodedKey)
			return nil, fmt.Errorf("%s", errorMsg)
		}

		// Success! Inform the user and exit.
		return nil, fmt.Errorf("[SETUP] JWT_SECRET was missing. A new secret has been generated and saved to your .env file. Please restart the application")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets a boolean environment variable with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvAsInt gets an integer environment variable with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
