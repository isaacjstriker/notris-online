package main

import (
	"log"
	"time"

	"github.com/isaacjstriker/devware/internal/api"
	"github.com/isaacjstriker/devware/internal/config"
	"github.com/isaacjstriker/devware/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Could not load configuration: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Could not connect to the database: %v", err)
	}
	defer db.Close()

	if err := db.CreateTables(); err != nil {
		log.Fatalf("[FATAL] Could not create database tables: %v", err)
	}

	// Start automatic cleanup routine
	startCleanupScheduler(db)

	server := api.NewAPIServer(cfg, db)
	server.Start()
}

// startCleanupScheduler runs periodic cleanup of inactive multiplayer rooms
func startCleanupScheduler(db *database.DB) {
	ticker := time.NewTicker(2 * time.Hour)

	go func() {
		log.Println("[INFO] Starting multiplayer room cleanup scheduler")

		runCleanup(db)

		for range ticker.C {
			runCleanup(db)
		}
	}()
}

func runCleanup(db *database.DB) {
	log.Println("[INFO] Running scheduled multiplayer room cleanup...")

	// Clean up rooms older than 4 hours
	maxAge := 4 * time.Hour

	roomsDeleted, err := db.CleanupInactiveRooms(maxAge)
	if err != nil {
		log.Printf("[ERROR] Failed to cleanup inactive rooms: %v", err)
		return
	}

	if len(roomsDeleted) > 0 {
		log.Printf("[INFO] Successfully cleaned up %d inactive rooms: %v", len(roomsDeleted), roomsDeleted)
	} else {
		log.Println("[INFO] No inactive rooms found to cleanup")
	}
}
