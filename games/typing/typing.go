package typing

import (
    "github.com/isaacjstriker/devware/internal/auth"
    "github.com/isaacjstriker/devware/internal/database"
    "github.com/isaacjstriker/devware/games"
)

// TypingGame implements the Game interface
type TypingGame struct{}

// NewTypingGame creates a new typing game instance
func NewTypingGame() *TypingGame {
    return &TypingGame{}
}

func (tg *TypingGame) GetName() string {
    return "Typing Speed Challenge"
}

func (tg *TypingGame) GetDescription() string {
    return "Test your typing speed and accuracy with random words"
}

func (tg *TypingGame) GetDifficulty() int {
    return 3 // Medium difficulty
}

func (tg *TypingGame) IsAvailable() bool {
    // Check if Lua files exist, keyboard library available, etc.
    return true
}

func (tg *TypingGame) Play(db *database.DB, authManager *auth.CLIAuth) *games.GameResult {
    // Run the existing game logic
    stats := playGame()
    
    // Convert to GameResult format
    result := &games.GameResult{
        GameName:  tg.GetName(),
        Score:     stats.Score,
        Duration:  stats.TotalTime,
        Accuracy:  stats.Accuracy,
        Perfect:   stats.Accuracy >= 95.0, // Define perfect as 95%+ accuracy
        Metadata: map[string]interface{}{
            "wpm":          stats.WPM,
            "words_typed":  stats.WordsTyped,
            "correct_words": stats.CorrectWords,
        },
    }
    
    // Calculate bonus for perfect games
    if result.Perfect {
        result.Bonus = int(float64(result.Score) * 0.2) // 20% bonus
        result.Score += result.Bonus
    }
    
    return result
}