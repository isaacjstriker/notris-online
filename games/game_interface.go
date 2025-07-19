package games

import (
    "github.com/isaacjstriker/devware/internal/auth"
    "github.com/isaacjstriker/devware/internal/database"
)

// GameResult represents the outcome of a single game
type GameResult struct {
    GameName     string  `json:"game_name"`
    Score        int     `json:"score"`
    Duration     float64 `json:"duration"`
    Accuracy     float64 `json:"accuracy"`
    Bonus        int     `json:"bonus"`
    Perfect      bool    `json:"perfect"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// ChallengeStats represents the overall challenge performance
type ChallengeStats struct {
    TotalScore    int           `json:"total_score"`
    GamesPlayed   int           `json:"games_played"`
    TotalDuration float64       `json:"total_duration"`
    AvgAccuracy   float64       `json:"avg_accuracy"`
    PerfectGames  int           `json:"perfect_games"`
    Results       []GameResult  `json:"results"`
}

// Game interface that all games must implement
type Game interface {
    // GetName returns the display name of the game
    GetName() string
    
    // GetDescription returns a brief description
    GetDescription() string
    
    // Play runs the game and returns the result
    Play(db *database.DB, authManager *auth.CLIAuth) *GameResult
    
    // GetDifficulty returns relative difficulty (1-10)
    GetDifficulty() int
    
    // IsAvailable checks if game can be played (dependencies, etc.)
    IsAvailable() bool
}