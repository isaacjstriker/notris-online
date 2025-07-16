package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type DB struct {
	*sql.DB
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
	var driverName string
	var dataSourceName string

	// Parse database URL to determine driver
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		driverName = "postgres"
		dataSourceName = databaseURL
	} else if strings.HasPrefix(databaseURL, "sqlite3://") {
		driverName = "sqlite3"
		dataSourceName = strings.TrimPrefix(databaseURL, "sqlite3://")
	} else {
		return nil, fmt.Errorf("unsupported database URL format: %s", databaseURL)
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// CreateTables creates the necessary database tables
func (db *DB) CreateTables() error {
	// SQLite-compatible table creation
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS game_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id),
			game_type TEXT NOT NULL,
			score INTEGER NOT NULL,
			additional_data TEXT,
			played_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_user_game 
		 ON game_scores(user_id, game_type)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_game_score 
		 ON game_scores(game_type, score DESC)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// CreateUser creates a new user in the database
func (db *DB) CreateUser(username, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, created_at
	`

	var user User
	err := db.QueryRow(query, username, email, passwordHash).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, string, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, last_login
		FROM users WHERE username = $1
	`

	var user User
	var passwordHash string
	err := db.QueryRow(query, username).Scan(
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
		VALUES ($1, $2, $3, $4)
	`

	var additionalDataJSON []byte
	if additionalData != nil {
		// Convert map to JSON - you'll need to import encoding/json
		// additionalDataJSON, _ = json.Marshal(additionalData)
	}

	_, err := db.Exec(query, userID, gameType, score, additionalDataJSON)
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
			gs.game_type,
			MAX(gs.score) as best_score,
			AVG(gs.score)::FLOAT as avg_score,
			COUNT(gs.id) as games_played,
			MAX(gs.played_at) as last_played
		FROM users u
		JOIN game_scores gs ON u.id = gs.user_id
		WHERE gs.game_type = $1
		GROUP BY u.id, u.username, gs.game_type
		ORDER BY best_score DESC
		LIMIT $2
	`

	rows, err := db.Query(query, gameType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.Username, &entry.GameType, &entry.BestScore,
			&entry.AvgScore, &entry.GamesPlayed, &entry.LastPlayed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
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
			$2 as game_type,
			COALESCE(MAX(gs.score), 0) as best_score,
			COALESCE(AVG(gs.score)::FLOAT, 0) as avg_score,
			COUNT(gs.id) as games_played,
			COALESCE(MAX(gs.played_at), NOW()) as last_played
		FROM users u
		LEFT JOIN game_scores gs ON u.id = gs.user_id AND gs.game_type = $2
		WHERE u.id = $1
		GROUP BY u.id, u.username
	`

	var entry LeaderboardEntry
	err := db.QueryRow(query, userID, gameType).Scan(
		&entry.Username, &entry.GameType, &entry.BestScore,
		&entry.AvgScore, &entry.GamesPlayed, &entry.LastPlayed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &entry, nil
}
