package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type DB struct {
	conn   *sql.DB
	dbType string // "postgres"
}

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login"`
}

type GameScore struct {
	ID             int                    `json:"id"`
	UserID         int                    `json:"user_id"`
	GameType       string                 `json:"game_type"`
	Score          int                    `json:"score"`
	AdditionalData map[string]interface{} `json:"additional_data"`
	PlayedAt       time.Time              `json:"played_at"`
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Username     string                 `json:"username"`
	GameType     string                 `json:"game_type"`
	BestScore    int                    `json:"best_score"`
	AvgScore     float64                `json:"avg_score"`
	GamesPlayed  int                    `json:"games_played"`
	LastPlayed   time.Time              `json:"last_played"`
	TotalLines   int                    `json:"total_lines,omitempty"`
	AvgPPM       float64                `json:"avg_ppm,omitempty"`
	BestTime     float64                `json:"best_time,omitempty"`
	TotalTime    float64                `json:"total_time,omitempty"`
	Achievements []string               `json:"achievements,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LeaderboardFilter represents filtering options for leaderboards
type LeaderboardFilter struct {
	TimePeriod string // "daily", "weekly", "monthly", "all"
	Category   string // "score", "speed", "efficiency", "endurance"
	UserID     *int   // Optional filter for specific user
}

// UserStats represents user statistics for a specific game
type UserStats struct {
	BestScore   int       `json:"best_score"`
	AvgScore    float64   `json:"avg_score"`
	GamesPlayed int       `json:"games_played"`
	LastPlayed  time.Time `json:"last_played"`
}

// Connect establishes a connection to the database.
// This function will now be the primary way to connect.
func Connect(dbURL string) (*DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Determine driver from URL prefix
	var driverName string
	if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
		driverName = "postgres"
	} else {
		// You can add other drivers here if needed in the future
		return nil, fmt.Errorf("unsupported database type for URL")
	}

	conn, err := sql.Open(driverName, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("[INFO] Successfully connected to %s database.\n", driverName)
	return &DB{conn: conn, dbType: driverName}, nil
}

// CreateTables creates the necessary database tables
func (db *DB) CreateTables() error {
	// PostgreSQL table creation
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS game_scores (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			game_type VARCHAR(50) NOT NULL,
			score INTEGER NOT NULL,
			metadata JSONB,
			played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS challenge_scores (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			total_score INTEGER NOT NULL DEFAULT 0,
			games_played INTEGER NOT NULL DEFAULT 0,
			avg_accuracy DECIMAL(5,2) NOT NULL DEFAULT 0.0,
			perfect_games INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_user_game ON game_scores(user_id, game_type)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_type_score ON game_scores(game_type, score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_challenge_scores_total ON challenge_scores(total_score DESC)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// Exec wrapper for convenience
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query wrapper for convenience
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow wrapper for convenience
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// CreateUser creates a new user in the database
func (db *DB) CreateUser(username, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	var user User
	err := db.conn.QueryRow(query, username, email, passwordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:        user.ID,
		Username:  username,
		Email:     email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, string, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, last_login
		FROM users WHERE username = $1
	`

	var user User
	var passwordHash string
	err := db.conn.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash,
		&user.CreatedAt, &user.LastLogin,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	return &user, passwordHash, nil
}

// SaveGameScore saves a game score to the database
func (db *DB) SaveGameScore(userID int, gameType string, score int, metadata map[string]interface{}) error {
	query := `
		INSERT INTO game_scores (user_id, game_type, score, metadata, played_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	var metadataValue interface{}
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataValue = metadataJSON // For PostgreSQL JSONB
	}

	_, err := db.conn.Exec(query, userID, gameType, score, metadataValue, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save game score: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the leaderboard for a specific game
func (db *DB) GetLeaderboard(gameType string, limit int) ([]LeaderboardEntry, error) {
	return db.GetFilteredLeaderboard(gameType, limit, LeaderboardFilter{TimePeriod: "all", Category: "score"})
}

// GetFilteredLeaderboard retrieves the leaderboard with advanced filtering
func (db *DB) GetFilteredLeaderboard(gameType string, limit int, filter LeaderboardFilter) ([]LeaderboardEntry, error) {
	// Build time filter condition
	timeCondition := ""
	switch filter.TimePeriod {
	case "daily":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 day'"
	case "weekly":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 week'"
	case "monthly":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 month'"
	default: // "all"
		timeCondition = ""
	}

	// Build category-specific ordering
	var orderBy string
	var selectFields string

	switch filter.Category {
	case "speed":
		// Fastest completion time (from metadata)
		selectFields = `
			u.username,
			MAX(gs.score) as best_score,
			AVG(gs.score) as avg_score,
			COUNT(gs.id) as games_played,
			MAX(gs.played_at) as last_played,
			MIN((gs.metadata->>'time_played')::float) as best_time,
			AVG((gs.metadata->>'time_played')::float) as total_time,
			SUM(COALESCE((gs.metadata->>'lines_cleared')::int, 0)) as total_lines,
			AVG(COALESCE((gs.metadata->>'ppm')::float, 0)) as avg_ppm
		`
		orderBy = "best_time ASC NULLS LAST"
	case "efficiency":
		// Highest pieces per minute
		selectFields = `
			u.username,
			MAX(gs.score) as best_score,
			AVG(gs.score) as avg_score,
			COUNT(gs.id) as games_played,
			MAX(gs.played_at) as last_played,
			MIN((gs.metadata->>'time_played')::float) as best_time,
			AVG((gs.metadata->>'time_played')::float) as total_time,
			SUM(COALESCE((gs.metadata->>'lines_cleared')::int, 0)) as total_lines,
			AVG(COALESCE((gs.metadata->>'ppm')::float, 0)) as avg_ppm
		`
		orderBy = "avg_ppm DESC NULLS LAST"
	case "endurance":
		// Longest game time
		selectFields = `
			u.username,
			MAX(gs.score) as best_score,
			AVG(gs.score) as avg_score,
			COUNT(gs.id) as games_played,
			MAX(gs.played_at) as last_played,
			MAX((gs.metadata->>'time_played')::float) as best_time,
			AVG((gs.metadata->>'time_played')::float) as total_time,
			SUM(COALESCE((gs.metadata->>'lines_cleared')::int, 0)) as total_lines,
			AVG(COALESCE((gs.metadata->>'ppm')::float, 0)) as avg_ppm
		`
		orderBy = "best_time DESC NULLS LAST"
	default: // "score"
		selectFields = `
			u.username,
			MAX(gs.score) as best_score,
			AVG(gs.score) as avg_score,
			COUNT(gs.id) as games_played,
			MAX(gs.played_at) as last_played,
			MIN((gs.metadata->>'time_played')::float) as best_time,
			AVG((gs.metadata->>'time_played')::float) as total_time,
			SUM(COALESCE((gs.metadata->>'lines_cleared')::int, 0)) as total_lines,
			AVG(COALESCE((gs.metadata->>'ppm')::float, 0)) as avg_ppm
		`
		orderBy = "best_score DESC"
	}

	userFilter := ""
	if filter.UserID != nil {
		userFilter = fmt.Sprintf("AND u.id = %d", *filter.UserID)
	}

	query := fmt.Sprintf(`
        SELECT %s
        FROM users u
        JOIN game_scores gs ON u.id = gs.user_id
        WHERE gs.game_type = $1 %s %s
        GROUP BY u.id, u.username
        ORDER BY %s
        LIMIT $2
    `, selectFields, timeCondition, userFilter, orderBy)

	rows, err := db.conn.Query(query, gameType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		var bestTime, totalTime, avgPPM sql.NullFloat64
		var totalLines sql.NullInt64

		err := rows.Scan(
			&entry.Username,
			&entry.BestScore,
			&entry.AvgScore,
			&entry.GamesPlayed,
			&entry.LastPlayed,
			&bestTime,
			&totalTime,
			&totalLines,
			&avgPPM,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}

		entry.GameType = gameType
		if bestTime.Valid {
			entry.BestTime = bestTime.Float64
		}
		if totalTime.Valid {
			entry.TotalTime = totalTime.Float64
		}
		if totalLines.Valid {
			entry.TotalLines = int(totalLines.Int64)
		}
		if avgPPM.Valid {
			entry.AvgPPM = avgPPM.Float64
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetUserStats retrieves statistics for a specific user and game
func (db *DB) GetUserStats(userID int, gameType string) (*LeaderboardEntry, error) {
	query := `
		SELECT 
			u.username,
			$1 as game_type,
			COALESCE(MAX(gs.score), 0) as best_score,
			COALESCE(AVG(gs.score), 0) as avg_score,
			COUNT(gs.id) as games_played,
			COALESCE(MAX(gs.played_at), CURRENT_TIMESTAMP) as last_played
		FROM users u
		LEFT JOIN game_scores gs ON u.id = gs.user_id AND gs.game_type = $2
		WHERE u.id = $3
		GROUP BY u.id, u.username
	`

	var entry LeaderboardEntry

	err := db.conn.QueryRow(query, gameType, gameType, userID).Scan(
		&entry.Username, &entry.GameType, &entry.BestScore,
		&entry.AvgScore, &entry.GamesPlayed, &entry.LastPlayed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &entry, nil
}

// GetUserAchievements calculates achievements for a user based on their game history
func (db *DB) GetUserAchievements(userID int, gameType string) ([]string, error) {
	var achievements []string

	// Query user's game statistics
	query := `
		SELECT 
			COUNT(*) as total_games,
			MAX(score) as best_score,
			AVG(score) as avg_score,
			SUM(COALESCE((metadata->>'lines_cleared')::int, 0)) as total_lines,
			MAX(COALESCE((metadata->>'lines_cleared')::int, 0)) as max_lines,
			AVG(COALESCE((metadata->>'ppm')::float, 0)) as avg_ppm,
			MAX(COALESCE((metadata->>'ppm')::float, 0)) as max_ppm,
			MIN((metadata->>'time_played')::float) as best_time,
			COUNT(CASE WHEN (metadata->>'tetrises')::int > 0 THEN 1 END) as games_with_tetris
		FROM game_scores 
		WHERE user_id = $1 AND game_type = $2
	`

	var totalGames, bestScore, maxLines, gamesWithTetris sql.NullInt64
	var avgScore, avgPPM, maxPPM, bestTime sql.NullFloat64
	var totalLines sql.NullInt64

	err := db.conn.QueryRow(query, userID, gameType).Scan(
		&totalGames, &bestScore, &avgScore, &totalLines, &maxLines,
		&avgPPM, &maxPPM, &bestTime, &gamesWithTetris,
	)
	if err != nil {
		return achievements, fmt.Errorf("failed to get user achievements: %w", err)
	}

	// Calculate achievements based on stats
	if totalGames.Valid {
		if totalGames.Int64 >= 1 {
			achievements = append(achievements, "First Game")
		}
		if totalGames.Int64 >= 10 {
			achievements = append(achievements, "Getting Started")
		}
		if totalGames.Int64 >= 50 {
			achievements = append(achievements, "Dedicated Player")
		}
		if totalGames.Int64 >= 100 {
			achievements = append(achievements, "Tetris Master")
		}
	}

	if bestScore.Valid {
		if bestScore.Int64 >= 10000 {
			achievements = append(achievements, "High Scorer")
		}
		if bestScore.Int64 >= 50000 {
			achievements = append(achievements, "Score Champion")
		}
		if bestScore.Int64 >= 100000 {
			achievements = append(achievements, "Legendary Score")
		}
	}

	if maxPPM.Valid && maxPPM.Float64 >= 30 {
		achievements = append(achievements, "Speed Demon")
	}
	if maxPPM.Valid && maxPPM.Float64 >= 50 {
		achievements = append(achievements, "Lightning Fast")
	}

	if gamesWithTetris.Valid && gamesWithTetris.Int64 > 0 {
		achievements = append(achievements, "First Tetris")
	}
	if gamesWithTetris.Valid && gamesWithTetris.Int64 >= 5 {
		achievements = append(achievements, "Tetris Expert")
	}

	if totalLines.Valid && totalLines.Int64 >= 100 {
		achievements = append(achievements, "Line Clearer")
	}
	if totalLines.Valid && totalLines.Int64 >= 1000 {
		achievements = append(achievements, "Line Master")
	}

	return achievements, nil
}

// GetRecentGames gets recent games for activity feed
func (db *DB) GetRecentGames(gameType string, limit int) ([]GameScore, error) {
	query := `
		SELECT gs.id, gs.user_id, gs.game_type, gs.score, gs.metadata, gs.played_at, u.username
		FROM game_scores gs
		JOIN users u ON gs.user_id = u.id
		WHERE gs.game_type = $1
		ORDER BY gs.played_at DESC
		LIMIT $2
	`

	rows, err := db.conn.Query(query, gameType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent games: %w", err)
	}
	defer rows.Close()

	var games []GameScore
	for rows.Next() {
		var game GameScore
		var username string
		var metadataJSON []byte

		err := rows.Scan(
			&game.ID, &game.UserID, &game.GameType, &game.Score,
			&metadataJSON, &game.PlayedAt, &username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent game: %w", err)
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &game.AdditionalData); err != nil {
				game.AdditionalData = make(map[string]interface{})
			}
		}

		// Add username to metadata for easier access
		if game.AdditionalData == nil {
			game.AdditionalData = make(map[string]interface{})
		}
		game.AdditionalData["username"] = username

		games = append(games, game)
	}

	return games, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
