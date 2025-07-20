package main

import (
	"strings"
	"fmt"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/internal/config"
)

func testSupabaseConnection() {
    fmt.Println("🧪 Testing Supabase connection...")
    
    // Get the DATABASE_URL from your .env
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("❌ Failed to load config: %v\n", err)
        return
    }
    
    // Try to connect directly to PostgreSQL
    if strings.HasPrefix(cfg.DatabaseURL, "postgresql://") {
        db, err := database.Connect(cfg.DatabaseURL)
        if err != nil {
            fmt.Printf("❌ Failed to connect to Supabase: %v\n", err)
            return
        }
        defer db.Close()
        
        // Test with a simple query
        var version string
        err = db.QueryRow("SELECT version()").Scan(&version)
        if err != nil {
            fmt.Printf("❌ Failed to query Supabase: %v\n", err)
            return
        }
        
        fmt.Println("✅ Successfully connected to Supabase!")
        fmt.Printf("📋 PostgreSQL version: %s\n", version[:50]+"...")
        
        // Test table access
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
        if err != nil {
            fmt.Printf("⚠️  Warning: Could not check tables: %v\n", err)
        } else {
            fmt.Printf("📊 Found %d tables in public schema\n", count)
        }
        
    } else {
        fmt.Println("⚠️  DATABASE_URL is not a PostgreSQL connection string")
    }
}