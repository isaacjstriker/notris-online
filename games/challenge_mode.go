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

// RunChallenge plays all available games in a random order.
// It no longer calculates or saves aggregate challenge stats.
func (cm *ChallengeMode) RunChallenge(db *database.DB, authManager *auth.CLIAuth) {
	games := cm.registry.GetRandomOrder()

	if len(games) == 0 {
		fmt.Println("No games available for challenge mode!")
		return
	}

	fmt.Println("\n--- CHALLENGE MODE ACTIVATED! ---")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("You will play %d games in random order.\n", len(games))
	fmt.Println("Individual game scores will be saved if you are logged in.")
	fmt.Println(strings.Repeat("=", 50))

	totalScore := 0

	for i, game := range games {
		fmt.Printf("\n--- Game %d/%d: %s ---\n", i+1, len(games), game.GetName())
		fmt.Printf("Description: %s\n", game.GetDescription())
		fmt.Printf("Difficulty: %d/10\n", game.GetDifficulty())

		fmt.Println("\nPress Enter to start...")
		fmt.Scanln()

		// Play the game
		result := game.Play(db, authManager)
		if result != nil {
			totalScore += result.Score
			// Show individual game result
			fmt.Println("\n" + strings.Repeat("-", 40))
			fmt.Printf("Game Complete: %s\n", result.GameName)
			fmt.Printf("Score: %d\n", result.Score)
			fmt.Println(strings.Repeat("-", 40))
		}

		// Brief pause between games
		if i < len(games)-1 {
			fmt.Println("\nGet ready for the next game...")
			time.Sleep(3 * time.Second)
		}
	}

	fmt.Println("\n" + strings.Repeat("*", 25))
	fmt.Println("CHALLENGE COMPLETE!")
	fmt.Printf("Your combined score for this session was: %d\n", totalScore)
	fmt.Println(strings.Repeat("*", 25))
}
