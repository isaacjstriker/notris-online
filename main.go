package main

import (
	"fmt"
	"log"

	"github.com/isaacjstriker/devware/internal/api"
	"github.com/isaacjstriker/devware/internal/config"
	"github.com/isaacjstriker/devware/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Could not load configuration: %v", err)
	}

	// Initialize database connection (no fallback)
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		// In a cloud environment, if the DB isn't available, we should exit.
		log.Fatalf("[FATAL] Could not connect to the database: %v", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	if err := db.CreateTables(); err != nil {
		log.Fatalf("[FATAL] Could not create database tables: %v", err)
	}

	// Create and start the API server
	listenAddr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	server := api.NewAPIServer(listenAddr, db, cfg)

	// The server now runs indefinitely, so the CLI part is removed.
	// For game logic, we will later integrate it via WebSockets.
	server.Start()
}

// All the CLI-specific functions below this line can be removed.
// e.g., showUserProfile, showLeaderboard, etc.
