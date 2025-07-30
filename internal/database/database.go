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
	Username    string    `json:"username"`
	GameType    string    `json:"game_type"`
	BestScore   int       `json:"best_score"`
	AvgScore    float64   `json:"avg_score"`
	GamesPlayed int       `json:"games_played"`
	LastPlayed  time.Time `json:"last_played"`
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
	query := `
        SELECT 
            u.username,
            MAX(gs.score) as best_score,
            AVG(gs.score) as avg_score,
            COUNT(gs.id) as games_played,
            MAX(gs.played_at) as last_played
        FROM users u
        JOIN game_scores gs ON u.id = gs.user_id
        WHERE gs.game_type = $1
        GROUP BY u.id, u.username
        ORDER BY best_score DESC
        LIMIT $2
    `

	rows, err := db.conn.Query(query, gameType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry

		err := rows.Scan(
			&entry.Username,
			&entry.BestScore,
			&entry.AvgScore,
			&entry.GamesPlayed,
			&entry.LastPlayed, // PostgreSQL returns time.Time directly
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}

		entry.GameType = gameType
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

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
