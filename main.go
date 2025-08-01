package main

import (
	"log"

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

	server := api.NewAPIServer(cfg, db)
	server.Start()
}
