package games

import (
    "fmt"
    "strings"
    "time"

    "github.com/isaacjstriker/devware/internal/auth"
    "github.com/isaacjstriker/devware/internal/database"
)

// ChallengeMode manages the multi-game challenge
type ChallengeMode struct {
    registry *GameRegistry
}

// NewChallengeMode creates a new challenge mode
func NewChallengeMode(registry *GameRegistry) *ChallengeMode {
    return &ChallengeMode{
        registry: registry,
    }
}

// RunChallenge executes the full game challenge
func (cm *ChallengeMode) RunChallenge(db *database.DB, authManager *auth.CLIAuth) *ChallengeStats {
    games := cm.registry.GetRandomOrder()
    
    if len(games) == 0 {
        fmt.Println("❌ No games available for challenge mode!")
        return nil
    }
    
    fmt.Println("\n🎮 CHALLENGE MODE ACTIVATED!")
    fmt.Println(strings.Repeat("=", 50))
    fmt.Printf("🎯 You will play %d games in random order\n", len(games))
    fmt.Println("🏆 Your final score will be the sum of all games")
    fmt.Println("💎 Perfect games (95%+ accuracy) get 20% bonus!")
    fmt.Println(strings.Repeat("=", 50))
    
    stats := &ChallengeStats{
        Results: make([]GameResult, 0, len(games)),
    }
    
    startTime := time.Now()
    
    for i, game := range games {
        fmt.Printf("\n🎲 Game %d/%d: %s\n", i+1, len(games), game.GetName())
        fmt.Printf("📝 %s\n", game.GetDescription())
        fmt.Printf("⭐ Difficulty: %d/10\n", game.GetDifficulty())
        
        fmt.Println("\nPress Enter to start...")
        fmt.Scanln()
        
        // Play the game
        result := game.Play(db, authManager)
        if result != nil {
            stats.Results = append(stats.Results, *result)
            stats.TotalScore += result.Score
            stats.GamesPlayed++
            stats.TotalDuration += result.Duration
            stats.AvgAccuracy += result.Accuracy
            
            if result.Perfect {
                stats.PerfectGames++
            }
            
            // Show individual game result
            cm.showGameResult(result, i+1, len(games))
        }
        
        // Brief pause between games
        if i < len(games)-1 {
            fmt.Println("\n⏳ Get ready for the next game...")
            time.Sleep(2 * time.Second)
        }
    }
    
    stats.TotalDuration = time.Since(startTime).Seconds()
    if stats.GamesPlayed > 0 {
        stats.AvgAccuracy /= float64(stats.GamesPlayed)
    }
    
    // Show final results
    cm.showFinalResults(stats)
    
    // Save challenge score to database
    if db != nil && authManager != nil && authManager.GetSession().IsLoggedIn() {
        cm.saveChallengeScore(db, authManager, stats)
    }
    
    return stats
}

func (cm *ChallengeMode) showGameResult(result *GameResult, current, total int) {
    fmt.Println("\n" + strings.Repeat("=", 40))
    fmt.Printf("📊 Game %d/%d Complete: %s\n", current, total, result.GameName)
    fmt.Println(strings.Repeat("=", 40))
    fmt.Printf("🎯 Score: %d", result.Score)
    if result.Bonus > 0 {
        fmt.Printf(" (includes %d bonus!)", result.Bonus)
    }
    fmt.Printf("\n⏱️  Duration: %.1f seconds\n", result.Duration)
    fmt.Printf("🎯 Accuracy: %.1f%%\n", result.Accuracy)
    if result.Perfect {
        fmt.Println("💎 PERFECT GAME! Bonus applied!")
    }
    fmt.Println(strings.Repeat("=", 40))
}

func (cm *ChallengeMode) showFinalResults(stats *ChallengeStats) {
    fmt.Println("\n" + strings.Repeat("🎉", 20))
    fmt.Println("🏆 CHALLENGE COMPLETE! 🏆")
    fmt.Println(strings.Repeat("🎉", 20))
    
    fmt.Printf("🎮 Games Played: %d\n", stats.GamesPlayed)
    fmt.Printf("🎯 Total Score: %d points\n", stats.TotalScore)
    fmt.Printf("⏱️  Total Time: %.1f seconds\n", stats.TotalDuration)
    fmt.Printf("📊 Average Accuracy: %.1f%%\n", stats.AvgAccuracy)
    fmt.Printf("💎 Perfect Games: %d/%d\n", stats.PerfectGames, stats.GamesPlayed)
    
    if stats.PerfectGames == stats.GamesPlayed && stats.GamesPlayed > 1 {
        fmt.Println("🔥 FLAWLESS VICTORY! All games perfect!")
    }
    
    fmt.Println("\n📋 Game Breakdown:")
    for i, result := range stats.Results {
        perfectIcon := ""
        if result.Perfect {
            perfectIcon = " 💎"
        }
        fmt.Printf("  %d. %s: %d points (%.1f%%)%s\n", 
            i+1, result.GameName, result.Score, result.Accuracy, perfectIcon)
    }
    
    fmt.Println(strings.Repeat("=", 50))
}

func (cm *ChallengeMode) saveChallengeScore(db *database.DB, authManager *auth.CLIAuth, stats *ChallengeStats) {
    session := authManager.GetSession().GetCurrentSession()
    if session == nil {
        return
    }
    
    // Save to database (you'll need to add this method to your database package)
    err := db.SaveChallengeScore(session.UserID, stats)
    if err != nil {
        fmt.Printf("⚠️  Failed to save challenge score: %v\n", err)
    } else {
        fmt.Println("✅ Challenge score saved!")
    }
}