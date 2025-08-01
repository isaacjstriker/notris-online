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

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, reading from environment")
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppName:     getEnv("APP_NAME", "Dev Ware"),
		Debug:       getEnvAsBool("DEBUG", false),
		ServerPort:  getEnvAsInt("SERVER_PORT", 8080),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.JWTSecret == "" {
		newKey := make([]byte, 32)
		if _, err := rand.Read(newKey); err != nil {
			return nil, fmt.Errorf("failed to generate a new JWT key: %w", err)
		}
		encodedKey := base64.StdEncoding.EncodeToString(newKey)

		envFilePath := ".env"
		f, err := os.OpenFile(envFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			errorMsg := fmt.Sprintf(`
FATAL: JWT_SECRET is not set and I couldn't write to the .env file.
Error: %v

Please create a .env file and add the following line:

JWT_SECRET=%s

`, err, encodedKey)
			return nil, fmt.Errorf("%s", errorMsg)
		}
		defer f.Close()

		newLine := fmt.Sprintf("\nJWT_SECRET=%s\n", encodedKey)
		if _, err := f.WriteString(newLine); err != nil {
			errorMsg := fmt.Sprintf(`
FATAL: JWT_SECRET is not set and I failed to write to the .env file.
Error: %v

Please add the following line to your .env file:

JWT_SECRET=%s

`, err, encodedKey)
			return nil, fmt.Errorf("%s", errorMsg)
		}

		return nil, fmt.Errorf("[SETUP] JWT_SECRET was missing. A new secret has been generated and saved to your .env file. Please restart the application")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
