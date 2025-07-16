package typing

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	lua "github.com/yuin/gopher-lua"

	"github.com/isaacjstriker/devware/internal/auth"
	"github.com/isaacjstriker/devware/internal/database"
)

type GameStats struct {
	WordsTyped   int
	CorrectWords int
	TotalTime    float64
	WPM          float64
	Accuracy     float64
	Score        int
}

func RunWithAuth(db *database.DB, authManager *auth.CLIAuth) {
	// Check if user wants to save scores
	saveScores := false
	if authManager.GetSession().IsLoggedIn() {
		saveScores = true
		fmt.Printf("\n🎯 Starting Typing Game for %s\n", authManager.GetSession().GetUserInfo())
	} else {
		fmt.Println("\n🎯 Starting Typing Game (Guest Mode)")
		fmt.Println("💡 Tip: Login to save your high scores!")
	}

	stats := playGame()

	// Display results
	displayResults(stats)

	// Save score if user is authenticated
	if saveScores && db != nil {
		session := authManager.GetSession().GetCurrentSession()
		if session != nil {
			err := saveGameScore(db, session.UserID, stats)
			if err != nil {
				fmt.Printf("⚠️  Warning: Could not save score: %v\n", err)
			} else {
				fmt.Println("✅ Score saved to your profile!")

				// Show personal best
				userStats, err := db.GetUserStats(session.UserID, "typing")
				if err == nil {
					fmt.Printf("🏆 Your best score: %d\n", userStats.BestScore)
					fmt.Printf("📊 Games played: %d\n", userStats.GamesPlayed)
					fmt.Printf("📈 Average score: %.1f\n", userStats.AvgScore)
				}
			}
		}
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

func RunLegacy() {
	// Fallback for backward compatibility
	stats := playGame()
	displayResults(stats)
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

func playGame() *GameStats {
	L := lua.NewState()
	defer L.Close()

	// Load Lua script
	if err := L.DoFile("games/typing/typing_game.lua"); err != nil {
		fmt.Println("Lua error:", err)
		return &GameStats{}
	}

	// Get random words list
	L.Push(L.GetGlobal("get_random_words"))
	L.Push(lua.LNumber(10)) // Change number to change amount of words cycled
	if err := L.PCall(1, 1, nil); err != nil {
		fmt.Println("Lua error:", err)
		return &GameStats{}
	}
	wordsTable := L.Get(-1)
	L.Pop(1)

	words := []string{}
	if tbl, ok := wordsTable.(*lua.LTable); ok {
		tbl.ForEach(func(_, value lua.LValue) {
			words = append(words, value.String())
		})
	}

	// Shuffle the words slice
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to open keyboard:", err)
		return &GameStats{}
	}
	defer keyboard.Close()

	stats := &GameStats{}
	timer := time.Now()

	fmt.Println("\n⏰ Get ready! Type the words as they appear...")
	fmt.Println("You have 5 seconds per word. Good luck!")
	fmt.Println()

	// Send words to channel in a goroutine
	wordChannel := make(chan string, 1)
	wordIndex := 0

	// Start the game by sending the first word
	if len(words) > 0 {
		fmt.Printf("Type this word: %s\n", words[wordIndex])
		go func() {
			wordChannel <- words[wordIndex]
		}()
	}

	// Game loop
gameLoop:
	for wordIndex < len(words) {
		select {
		case word := <-wordChannel:
			// Get user input
			userInput := ""
			fmt.Print("> ")

			// Read user input character by character
			for {
				char, key, err := keyboard.GetKey()
				if err != nil {
					fmt.Println("Error reading keyboard:", err)
					break gameLoop
				}

				if key == keyboard.KeyEnter {
					fmt.Println() // New line after enter
					break
				} else if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
					if len(userInput) > 0 {
						userInput = userInput[:len(userInput)-1]
						fmt.Print("\b \b") // Backspace visual effect
					}
				} else if char != 0 {
					userInput += string(char)
					fmt.Print(string(char))
				}
			}

			stats.WordsTyped++

			// Check if input matches the word
			if strings.TrimSpace(userInput) == word {
				fmt.Println("✅ Correct!")
				stats.CorrectWords++

				// Move to next word
				wordIndex++
				if wordIndex < len(words) {
					fmt.Printf("\nType this word: %s\n", words[wordIndex])
					go func(w string) {
						wordChannel <- w
					}(words[wordIndex])
				}
			} else {
				fmt.Println("❌ Incorrect. Game over.")
				break gameLoop
			}

		case <-time.After(5 * time.Second):
			fmt.Println("\n⏰ Time's up! Game over.")
			break gameLoop
		}
	}

	if wordIndex >= len(words) {
		fmt.Println("\n🎉 Congratulations! You completed all words!")
	}

	stats.TotalTime = time.Since(timer).Seconds()

	// Calculate final stats
	if stats.TotalTime > 0 {
		stats.WPM = float64(stats.CorrectWords) / (stats.TotalTime / 60.0)
	}
	if stats.WordsTyped > 0 {
		stats.Accuracy = float64(stats.CorrectWords) / float64(stats.WordsTyped) * 100
	}
	stats.Score = int(stats.WPM * stats.Accuracy / 100 * 10) // Custom scoring formula

	return stats
}

func displayResults(stats *GameStats) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎯 TYPING GAME RESULTS")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📝 Words Attempted: %d\n", stats.WordsTyped)
	fmt.Printf("✅ Words Correct: %d\n", stats.CorrectWords)
	fmt.Printf("⏱️  Total Time: %.2f seconds\n", stats.TotalTime)
	fmt.Printf("⚡ Words Per Minute: %.1f WPM\n", stats.WPM)
	fmt.Printf("🎯 Accuracy: %.1f%%\n", stats.Accuracy)
	fmt.Printf("🏆 Final Score: %d points\n", stats.Score)
	fmt.Println(strings.Repeat("=", 50))
}

func saveGameScore(db *database.DB, userID int, stats *GameStats) error {
	additionalData := map[string]interface{}{
		"wpm":           stats.WPM,
		"accuracy":      stats.Accuracy,
		"words_typed":   stats.WordsTyped,
		"correct_words": stats.CorrectWords,
		"total_time":    stats.TotalTime,
	}

	return db.SaveGameScore(userID, "typing", stats.Score, additionalData)
}
