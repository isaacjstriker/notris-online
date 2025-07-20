package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/isaacjstriker/devware/internal/types" // Change this import

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // Keep SQLite driver for local development
)

type DB struct {
	conn   *sql.DB
	dbType string // "postgres" or "sqlite3"
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

type LeaderboardEntry struct {
	Username    string    `json:"username"`
	GameType    string    `json:"game_type"`
	BestScore   int       `json:"best_score"`
	AvgScore    float64   `json:"avg_score"`
	GamesPlayed int       `json:"games_played"`
	LastPlayed  time.Time `json:"last_played"`
}

// Connect establishes a connection to the database
func Connect(databaseURL string) (*DB, error) {
	var db *sql.DB
	var err error
	var dbType string

	if strings.HasPrefix(databaseURL, "postgresql://") || strings.HasPrefix(databaseURL, "postgres://") {
		// PostgreSQL connection
		db, err = sql.Open("postgres", databaseURL)
		dbType = "postgres"
	} else {
		// SQLite connection (for local development)
		db, err = sql.Open("sqlite3", databaseURL)
		dbType = "sqlite3"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn:   db,
		dbType: dbType,
	}, nil
}

// CreateTables creates the necessary database tables
func (db *DB) CreateTables() error {
	var queries []string

	if db.dbType == "postgres" {
		// PostgreSQL table creation
		queries = []string{
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
	} else {
		// SQLite table creation (for local development)
		queries = []string{
			`CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL,
				email TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				last_login DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS game_scores (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER,
				game_type TEXT NOT NULL,
				score INTEGER NOT NULL,
				metadata TEXT,
				played_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users (id)
			)`,
			`CREATE TABLE IF NOT EXISTS challenge_scores (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER,
				total_score INTEGER NOT NULL DEFAULT 0,
				games_played INTEGER NOT NULL DEFAULT 0,
				avg_accuracy REAL NOT NULL DEFAULT 0.0,
				perfect_games INTEGER NOT NULL DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users (id)
			)`,
		}
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
		VALUES (?, ?, ?)
	`

	result, err := db.conn.Exec(query, username, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	return &User{
		ID:        int(id),
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, string, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, last_login
		FROM users WHERE username = ?
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
func (db *DB) SaveGameScore(userID int, gameType string, score int, additionalData map[string]interface{}) error {
	query := `
		INSERT INTO game_scores (user_id, game_type, score, additional_data)
		VALUES (?, ?, ?, ?)
	`

	var additionalDataJSON []byte
	if additionalData != nil {
		var err error
		additionalDataJSON, err = json.Marshal(additionalData)
		if err != nil {
			return fmt.Errorf("failed to marshal additional data: %w", err)
		}
	}

	_, err := db.conn.Exec(query, userID, gameType, score, string(additionalDataJSON))
	if err != nil {
		return fmt.Errorf("failed to save game score: %w", err)
	}

	return nil
}

// SubmitScore submits a game score to the database
func (db *DB) SubmitScore(userID int, gameType string, score int, metadata map[string]interface{}) error {
	var query string
	var args []interface{}

	if db.dbType == "postgres" {
		query = `INSERT INTO game_scores (user_id, game_type, score, metadata) VALUES ($1, $2, $3, $4)`

		var metadataJSON []byte
		var err error
		if metadata != nil {
			metadataJSON, err = json.Marshal(metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
		}

		args = []interface{}{userID, gameType, score, metadataJSON}
	} else {
		query = `INSERT INTO game_scores (user_id, game_type, score, metadata) VALUES (?, ?, ?, ?)`

		var metadataJSON string
		if metadata != nil {
			jsonBytes, err := json.Marshal(metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataJSON = string(jsonBytes)
		}

		args = []interface{}{userID, gameType, score, metadataJSON}
	}

	_, err := db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to submit score: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the leaderboard for a specific game
func (db *DB) GetLeaderboard(gameType string, limit int) ([]LeaderboardEntry, error) {
	var query string

	if db.dbType == "postgres" {
		query = `
            SELECT 
                u.username,
                $2 as game_type,
                MAX(gs.score) as best_score,
                AVG(gs.score) as avg_score,
                COUNT(gs.id) as games_played,
                MAX(gs.played_at) as last_played
            FROM users u
            JOIN game_scores gs ON u.id = gs.user_id
            WHERE gs.game_type = $2
            GROUP BY u.id, u.username
            ORDER BY best_score DESC
            LIMIT $1
        `
	} else {
		query = `
            SELECT 
                u.username,
                ? as game_type,
                MAX(gs.score) as best_score,
                AVG(CAST(gs.score AS REAL)) as avg_score,
                COUNT(gs.id) as games_played,
                MAX(gs.played_at) as last_played
            FROM users u
            JOIN game_scores gs ON u.id = gs.user_id
            WHERE gs.game_type = ?
            GROUP BY u.id, u.username
            ORDER BY best_score DESC
            LIMIT ?
        `
	}

	var rows *sql.Rows
	var err error

	if db.dbType == "postgres" {
		rows, err = db.conn.Query(query, limit, gameType)
	} else {
		rows, err = db.conn.Query(query, gameType, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		var lastPlayed time.Time

		err := rows.Scan(
			&entry.Username, &entry.GameType, &entry.BestScore,
			&entry.AvgScore, &entry.GamesPlayed, &lastPlayed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}

		entry.LastPlayed = lastPlayed
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetUserStats retrieves statistics for a specific user and game
func (db *DB) GetUserStats(userID int, gameType string) (*LeaderboardEntry, error) {
	query := `
		SELECT 
			u.username,
			? as game_type,
			COALESCE(MAX(gs.score), 0) as best_score,
			COALESCE(AVG(CAST(gs.score AS REAL)), 0) as avg_score,
			COUNT(gs.id) as games_played,
			COALESCE(MAX(gs.played_at), datetime('now')) as last_played
		FROM users u
		LEFT JOIN game_scores gs ON u.id = gs.user_id AND gs.game_type = ?
		WHERE u.id = ?
		GROUP BY u.id, u.username
	`

	var entry LeaderboardEntry
	var lastPlayedStr string

	err := db.conn.QueryRow(query, gameType, gameType, userID).Scan(
		&entry.Username, &entry.GameType, &entry.BestScore,
		&entry.AvgScore, &entry.GamesPlayed, &lastPlayedStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Parse the date string into time.Time
	if lastPlayedStr != "" {
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			time.RFC3339,
		}

		var parsed bool
		for _, format := range formats {
			if t, err := time.Parse(format, lastPlayedStr); err == nil {
				entry.LastPlayed = t
				parsed = true
				break
			}
		}

		if !parsed {
			entry.LastPlayed = time.Now()
		}
	} else {
		entry.LastPlayed = time.Now()
	}

	return &entry, nil
}

// SaveChallengeScore saves a challenge score to the database
func (db *DB) SaveChallengeScore(userID int, stats *types.ChallengeStats) error {
	query := `
        INSERT INTO challenge_scores (
            user_id, total_score, games_played, total_duration, 
            avg_accuracy, perfect_games, results_json
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	resultsJSON, err := json.Marshal(stats.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	_, err = db.conn.Exec(query,
		userID,
		stats.TotalScore,
		stats.GamesPlayed,
		stats.TotalDuration,
		stats.AvgAccuracy,
		stats.PerfectGames,
		string(resultsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to save challenge score: %w", err)
	}

	return nil
}

// GetUserChallengeScores gets challenge scores for a user
func (db *DB) GetUserChallengeScores(userID int, limit int) ([]map[string]interface{}, error) {
	query := `
        SELECT id, total_score, games_played, total_duration, 
               avg_accuracy, perfect_games, results_json, created_at
        FROM challenge_scores 
        WHERE user_id = ? 
        ORDER BY created_at DESC 
        LIMIT ?
    `

	rows, err := db.conn.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user challenge scores: %w", err)
	}
	defer rows.Close()

	var scores []map[string]interface{}
	for rows.Next() {
		var id, totalScore, gamesPlayed, perfectGames int
		var totalDuration, avgAccuracy float64
		var resultsJSON string
		var createdAt time.Time

		err := rows.Scan(&id, &totalScore, &gamesPlayed, &totalDuration,
			&avgAccuracy, &perfectGames, &resultsJSON, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge score: %w", err)
		}

		score := map[string]interface{}{
			"id":             id,
			"total_score":    totalScore,
			"games_played":   gamesPlayed,
			"total_duration": totalDuration,
			"avg_accuracy":   avgAccuracy,
			"perfect_games":  perfectGames,
			"results_json":   resultsJSON,
			"created_at":     createdAt,
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetTopChallengeScores gets the top challenge scores across all users
func (db *DB) GetTopChallengeScores(limit int) ([]map[string]interface{}, error) {
	query := `
        SELECT cs.id, cs.total_score, cs.games_played, cs.total_duration,
               cs.avg_accuracy, cs.perfect_games, cs.created_at, u.username
        FROM challenge_scores cs
        JOIN users u ON cs.user_id = u.id
        ORDER BY cs.total_score DESC 
        LIMIT ?
    `

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top challenge scores: %w", err)
	}
	defer rows.Close()

	var scores []map[string]interface{}
	for rows.Next() {
		var id, totalScore, gamesPlayed, perfectGames int
		var totalDuration, avgAccuracy float64
		var createdAt time.Time
		var username string

		err := rows.Scan(&id, &totalScore, &gamesPlayed, &totalDuration,
			&avgAccuracy, &perfectGames, &createdAt, &username)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge score: %w", err)
		}

		score := map[string]interface{}{
			"id":             id,
			"total_score":    totalScore,
			"games_played":   gamesPlayed,
			"total_duration": totalDuration,
			"avg_accuracy":   avgAccuracy,
			"perfect_games":  perfectGames,
			"created_at":     createdAt,
			"username":       username,
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
