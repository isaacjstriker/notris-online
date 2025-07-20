package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/isaacjstriker/devware/games"
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

	//Test connection to database first
	testSupabaseConnection()

	// Initialize database connection with fallback
	db, err := database.ConnectWithFallback(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to any database: %v", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Add test data if database is empty
	if err := db.CreateTestData(); err != nil {
		fmt.Printf("ℹ️  Test data: %v\n", err)
	}

	fmt.Println("✅ Database connected and tables verified.")

	var authManager *auth.CLIAuth

	// Initialize authentication if database is available
	if db != nil {
		authManager = auth.NewCLIAuth(db)
		fmt.Println("🔐 Authentication system initialized.")
	}

	// Main application loop
	for {
		// Build menu items step by step
		var menuItems []ui.MenuItem

		// Always available items
		menuItems = append(menuItems,
			ui.MenuItem{Label: "🎲 Challenge Mode (All Games)", Value: "challenge"},
			ui.MenuItem{Label: "🎯 Typing Speed Challenge", Value: "typing"},
		)

		// User-specific items
		if authManager != nil && authManager.GetSession().IsLoggedIn() {
			session := authManager.GetSession().GetCurrentSession()
			userDisplayName := "User"
			if session != nil {
				userDisplayName = session.Username
			}

			menuItems = []ui.MenuItem{
				{Label: fmt.Sprintf("👤 %s", userDisplayName), Value: "user_info"},
				{Label: "🎲 Challenge Mode (All Games)", Value: "challenge"},
				{Label: "🎯 Typing Speed Challenge", Value: "typing"},
				{Label: "🏆 View Leaderboards", Value: "leaderboard"},
				{Label: "🔄 Authentication", Value: "auth"},
			}
		} else if authManager != nil {
			menuItems = append(menuItems,
				ui.MenuItem{Label: "👤 Login / Register", Value: "auth"},
				ui.MenuItem{Label: "🏆 View Leaderboards", Value: "leaderboard"},
			)
		}

		// Always at the end
		menuItems = append(menuItems,
			ui.MenuItem{Label: "⚙️  Settings", Value: "settings"},
			ui.MenuItem{Label: "❌ Exit", Value: "exit"},
		)

		// Create and show menu
		menu := ui.NewMenu("Main Menu - Select Your Adventure", menuItems)
		choice := menu.Show()

		switch choice {
		case "typing":
			typingGame := typing.NewTypingGame()
			if authManager != nil {
				typingGame.Play(db, authManager)
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

		case "challenge":
			gameRegistry := games.NewGameRegistry()
			gameRegistry.RegisterGame(typing.NewTypingGame())
			// Register other games as you create them

			challengeMode := games.NewChallengeMode(gameRegistry)
			challengeMode.RunChallenge(db, authManager)

		default:
			fmt.Println("Invalid selection. Please try again.")
		}

		// Debug: Show login status
		if authManager != nil && authManager.GetSession().IsLoggedIn() {
			session := authManager.GetSession().GetCurrentSession()
			fmt.Printf("🔐 Debug: Logged in as %s (ID: %d)\n", session.Username, session.UserID)
		} else {
			fmt.Println("❌ Debug: Not logged in")
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
	// For now, we'll show typing game by default
	// Later we can expand this to show a game selection menu

	availableGames := map[string]string{
		"typing": "🎯 Typing Speed Challenge",
		// Add more games here as you create them:
		// "memory": "🧠 Memory Challenge",
		// "math":   "🔢 Math Speed Test",
	}

	// For now, let's show typing game leaderboard
	// In the future, we'll add a submenu here
	gameType := "typing"
	gameName := availableGames[gameType]

	fmt.Println("\n" + strings.Repeat("🏆", 25))
	fmt.Printf("         LEADERBOARDS - %s\n", gameName)
	fmt.Println(strings.Repeat("🏆", 25))

	// Get leaderboard data
	entries, err := db.GetLeaderboard(gameType, 15) // Show top 15
	if err != nil {
		fmt.Printf("❌ Error loading leaderboard: %v\n", err)
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return
	}

	if len(entries) == 0 {
		fmt.Println("\n📝 No scores recorded yet!")
		fmt.Println("🎮 Be the first to play and set a record!")
		fmt.Println("💡 Log in and play some games to see your scores here.")
	} else {
		fmt.Printf("\n📊 Showing Top %d Players:\n", len(entries))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("%-4s | %-15s | %-10s | %-6s | %-8s | %s\n",
			"Rank", "Player", "Best Score", "Games", "Avg", "Last Played")
		fmt.Println(strings.Repeat("-", 70))

		for i, entry := range entries {
			// Format the last played date
			lastPlayed := entry.LastPlayed.Format("Jan 02")

			// Add medal emojis for top 3
			rankDisplay := fmt.Sprintf("%d", i+1)
			switch i {
			case 0:
				rankDisplay = "🥇"
			case 1:
				rankDisplay = "🥈"
			case 2:
				rankDisplay = "🥉"
			}

			fmt.Printf("%-4s | %-15s | %-10d | %-6d | %-8.1f | %s\n",
				rankDisplay,
				truncateString(entry.Username, 15),
				entry.BestScore,
				entry.GamesPlayed,
				entry.AvgScore,
				lastPlayed)
		}

		fmt.Println(strings.Repeat("=", 70))

		// Show some interesting stats
		showLeaderboardStats(entries)
	}

	fmt.Println("\n💡 Future: We'll add a menu to select different game leaderboards!")
	fmt.Println("🎮 For now, only Typing Speed Challenge is available.")
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

// Helper function to show interesting leaderboard statistics
func showLeaderboardStats(entries []database.LeaderboardEntry) {
	if len(entries) == 0 {
		return
	}

	// Calculate some interesting stats
	totalGames := 0
	totalScore := 0
	for _, entry := range entries {
		totalGames += entry.GamesPlayed
		totalScore += entry.BestScore
	}

	avgBestScore := float64(totalScore) / float64(len(entries))

	fmt.Println("\n📈 Community Stats:")
	fmt.Printf("   🎮 Total Games Played: %d\n", totalGames)
	fmt.Printf("   👥 Active Players: %d\n", len(entries))
	fmt.Printf("   📊 Average Best Score: %.1f\n", avgBestScore)

	if len(entries) > 0 {
		fmt.Printf("   🏆 Highest Score: %d (by %s)\n", entries[0].BestScore, entries[0].Username)
	}
}

// Helper function to truncate long usernames
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
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