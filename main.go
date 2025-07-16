package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/isaacjstriker/devware/games/typing"
	"github.com/isaacjstriker/devware/internal/auth"
	"github.com/isaacjstriker/devware/internal/config"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection (if configured)
	var db *database.DB
	var authManager *auth.CLIAuth

	if cfg.DatabaseURL != "" && cfg.DatabaseURL != "disabled" {
		fmt.Println("🔗 Connecting to database...")
		db, err = database.Connect(cfg.DatabaseURL)
		if err != nil {
			fmt.Printf("⚠️  Warning: Database connection failed: %v\n", err)
			fmt.Println("🔄 Continuing in offline mode...")
		} else {
			fmt.Println("✅ Database connected successfully!")

			// Create tables if they don't exist
			if err := db.CreateTables(); err != nil {
				fmt.Printf("⚠️  Warning: Failed to create tables: %v\n", err)
			}

			// Initialize authentication
			authManager = auth.NewCLIAuth(db)
		}
	} else {
		fmt.Println("📝 Running in offline mode (no database configured)")
	}

	// Main application loop
	for {
		// Create menu items based on available features
		var menuItems []ui.MenuItem

		if authManager != nil && authManager.GetSession().IsLoggedIn() {
			// Show user info in menu when logged in
			userInfo := authManager.GetSession().GetUserInfo()
			menuItems = []ui.MenuItem{
				{Label: fmt.Sprintf("👤 %s", userInfo), Value: "user_info"},
				{Label: "🎯 Typing Speed Challenge", Value: "typing"},
				{Label: "🏆 View Leaderboards", Value: "leaderboard"},
				{Label: "🔄 Authentication", Value: "auth"},
				{Label: "⚙️  Settings", Value: "settings"},
				{Label: "❌ Exit", Value: "exit"},
			}
		} else {
			menuItems = []ui.MenuItem{
				{Label: "🎯 Typing Speed Challenge", Value: "typing"},
			}

			// Add auth option only if database is available
			if authManager != nil {
				menuItems = append(menuItems, ui.MenuItem{Label: "👤 Login / Register", Value: "auth"})
				menuItems = append(menuItems, ui.MenuItem{Label: "🏆 View Leaderboards", Value: "leaderboard"})
			}

			menuItems = append(menuItems,
				ui.MenuItem{Label: "⚙️  Settings", Value: "settings"},
				ui.MenuItem{Label: "❌ Exit", Value: "exit"},
			)
		}

		// Create and show menu
		menu := ui.NewMenu("Main Menu - Select Your Adventure", menuItems)
		choice := menu.Show()

		switch choice {
		case "typing":
			if authManager != nil {
				typing.RunWithAuth(db, authManager)
			}

		case "auth":
			if authManager != nil {
				authManager.ShowAuthMenu()
			} else {
				fmt.Println("\n⚠️  Authentication not available (no database connection)")
				fmt.Println("Press Enter to continue...")
				fmt.Scanln()
			}

		case "user_info":
			if authManager != nil && authManager.GetSession().IsLoggedIn() {
				showUserProfile(db, authManager)
			}

		case "leaderboard":
			if db != nil {
				showLeaderboard(db)
			} else {
				fmt.Println("\n⚠️  Leaderboards not available (no database connection)")
				fmt.Println("Press Enter to continue...")
				fmt.Scanln()
			}

		case "settings":
			showSettings(cfg)

		case "exit":
			fmt.Println("\n👋 Thanks for playing Dev Ware!")
			fmt.Println("💝 Come back soon for more games!")
			if authManager != nil && authManager.GetSession().IsLoggedIn() {
				fmt.Printf("🔐 %s will remain logged in for next time.\n", authManager.GetSession().GetCurrentSession().Username)
			}
			return

		default:
			fmt.Println("Invalid selection. Please try again.")
		}
	}
}

func showUserProfile(db *database.DB, authManager *auth.CLIAuth) {
	session := authManager.GetSession().GetCurrentSession()
	if session == nil {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("👤 User Profile: %s\n", session.Username)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📧 Email: %s\n", session.Email)
	fmt.Printf("🆔 User ID: %d\n", session.UserID)

	// Get typing game stats
	if stats, err := db.GetUserStats(session.UserID, "typing"); err == nil {
		fmt.Println("\n🎯 Typing Game Statistics:")
		fmt.Printf("   🏆 Best Score: %d points\n", stats.BestScore)
		fmt.Printf("   🎮 Games Played: %d\n", stats.GamesPlayed)
		fmt.Printf("   📊 Average Score: %.1f points\n", stats.AvgScore)
		fmt.Printf("   ⏰ Last Played: %s\n", stats.LastPlayed.Format("2006-01-02 15:04"))
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func showLeaderboard(db *database.DB) {
	fmt.Println("\n🏆 LEADERBOARDS")
	fmt.Println("================")

	// Show typing game leaderboard
	entries, err := db.GetLeaderboard("typing", 10)
	if err != nil {
		fmt.Printf("Error loading leaderboard: %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	if len(entries) == 0 {
		fmt.Println("No scores recorded yet. Be the first to play!")
	} else {
		fmt.Println("\n🎯 Typing Speed Challenge - Top 10")
		fmt.Println("Rank | Player          | Best Score | Games | Avg Score")
		fmt.Println(strings.Repeat("=", 55))

		for i, entry := range entries {
			fmt.Printf("%-4d | %-15s | %-10d | %-5d | %.1f\n",
				i+1, entry.Username, entry.BestScore,
				entry.GamesPlayed, entry.AvgScore)
		}
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

func showSettings(cfg *config.Config) {
	fmt.Println("\n⚙️  SETTINGS")
	fmt.Println("============")
	fmt.Printf("App Name: %s\n", cfg.AppName)
	fmt.Printf("Debug Mode: %t\n", cfg.Debug)

	if cfg.DatabaseURL == "" || cfg.DatabaseURL == "disabled" {
		fmt.Println("Database: Disabled (Offline Mode)")
	} else {
		fmt.Printf("Database: Connected (%s)\n", cfg.DatabaseURL)
	}

	fmt.Printf("Server: %s:%d\n", cfg.ServerHost, cfg.ServerPort)

	fmt.Println("\n💡 Tip: Create a .env file to customize these settings")
	fmt.Println("(Copy .env.example and modify as needed)")

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}
