package types

// Remove these problematic imports:
// "github.com/isaacjstriker/devware/internal/auth"
// "github.com/isaacjstriker/devware/internal/database"

// GameResult represents the outcome of a single game
type GameResult struct {
    GameName     string                 `json:"game_name"`
    Score        int                    `json:"score"`
    Duration     float64                `json:"duration"`
    Accuracy     float64                `json:"accuracy"`
    Perfect      bool                   `json:"perfect"`
    Bonus        int                    `json:"bonus"`
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

// GameStats represents the statistics from a typing game
type GameStats struct {
    Score        int     `json:"score"`
    WPM          float64 `json:"wpm"`
    Accuracy     float64 `json:"accuracy"`
    WordsTyped   int     `json:"words_typed"`
    CorrectWords int     `json:"correct_words"`
    TotalTime    float64 `json:"total_time"`
}

// Game interface that all games must implement
// Note: We can't import auth/database here, so we use interfaces
type Game interface {
    // GetName returns the display name of the game
    GetName() string
    
    // GetDescription returns a brief description
    GetDescription() string
    
    // Play runs the game and returns the result
    // We'll use interface{} temporarily to avoid import cycle
    Play(db interface{}, authManager interface{}) *GameResult
    
    // GetDifficulty returns relative difficulty (1-10)
    GetDifficulty() int
    
    // IsAvailable checks if game can be played
    IsAvailable() bool
}