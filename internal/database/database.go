package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type DB struct {
	conn   *sql.DB
	dbType string
}

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
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

type MultiplayerRoom struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	GameType   string                 `json:"game_type"`
	MaxPlayers int                    `json:"max_players"`
	Status     string                 `json:"status"`
	CreatedBy  int                    `json:"created_by"`
	CreatedAt  time.Time              `json:"created_at"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
	Players    []MultiplayerPlayer    `json:"players,omitempty"`
	Spectators []int                  `json:"spectators,omitempty"`
}

type MultiplayerPlayer struct {
	UserID     int                    `json:"user_id"`
	Username   string                 `json:"username"`
	Position   int                    `json:"position"`
	Score      int                    `json:"score"`
	Status     string                 `json:"status"`
	JoinedAt   time.Time              `json:"joined_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	GameState  map[string]interface{} `json:"game_state,omitempty"`
	IsReady    bool                   `json:"is_ready"`
}

type MultiplayerGame struct {
	ID         string              `json:"id"`
	RoomID     string              `json:"room_id"`
	GameType   string              `json:"game_type"`
	Duration   int                 `json:"duration"`
	StartedAt  time.Time           `json:"started_at"`
	FinishedAt time.Time           `json:"finished_at"`
	Winner     *int                `json:"winner,omitempty"`
	Players    []MultiplayerPlayer `json:"players"`
}

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

type LeaderboardFilter struct {
	TimePeriod string
	Category   string
	UserID     *int
}

type UserStats struct {
	BestScore   int       `json:"best_score"`
	AvgScore    float64   `json:"avg_score"`
	GamesPlayed int       `json:"games_played"`
	LastPlayed  time.Time `json:"last_played"`
}

func Connect(dbURL string) (*DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	var driverName string
	if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
		driverName = "postgres"
	} else {
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

func (db *DB) CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
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
		`CREATE TABLE IF NOT EXISTS multiplayer_rooms (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			game_type VARCHAR(50) NOT NULL,
			max_players INTEGER NOT NULL DEFAULT 4,
			status VARCHAR(20) NOT NULL DEFAULT 'waiting',
			created_by INTEGER REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			finished_at TIMESTAMP,
			settings JSONB
		)`,
		`CREATE TABLE IF NOT EXISTS multiplayer_players (
			room_id VARCHAR(50) REFERENCES multiplayer_rooms(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			position INTEGER DEFAULT 0,
			score INTEGER DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'playing',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMP,
			game_state JSONB,
			is_ready BOOLEAN DEFAULT FALSE,
			PRIMARY KEY (room_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS multiplayer_games (
			id VARCHAR(50) PRIMARY KEY,
			room_id VARCHAR(50) REFERENCES multiplayer_rooms(id) ON DELETE CASCADE,
			game_type VARCHAR(50) NOT NULL,
			duration INTEGER NOT NULL,
			started_at TIMESTAMP NOT NULL,
			finished_at TIMESTAMP NOT NULL,
			winner INTEGER REFERENCES users(id),
			metadata JSONB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_user_game ON game_scores(user_id, game_type)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_type_score ON game_scores(game_type, score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_challenge_scores_total ON challenge_scores(total_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_multiplayer_rooms_status ON multiplayer_rooms(status, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_multiplayer_players_room ON multiplayer_players(room_id, joined_at)`,
		`CREATE INDEX IF NOT EXISTS idx_multiplayer_games_type ON multiplayer_games(game_type, finished_at DESC)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

func (db *DB) CreateUser(username, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	var user User
	err := db.conn.QueryRow(query, username, passwordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:        user.ID,
		Username:  username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (db *DB) GetUserByUsername(username string) (*User, string, error) {
	query := `
		SELECT id, username, password_hash, created_at, last_login
		FROM users WHERE username = $1
	`

	var user User
	var passwordHash string
	err := db.conn.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash,
		&user.CreatedAt, &user.LastLogin,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	return &user, passwordHash, nil
}

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
		metadataValue = metadataJSON
	}

	_, err := db.conn.Exec(query, userID, gameType, score, metadataValue, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save game score: %w", err)
	}

	return nil
}

func (db *DB) GetLeaderboard(gameType string, limit int) ([]LeaderboardEntry, error) {
	return db.GetFilteredLeaderboard(gameType, limit, LeaderboardFilter{TimePeriod: "all", Category: "score"})
}

func (db *DB) GetFilteredLeaderboard(gameType string, limit int, filter LeaderboardFilter) ([]LeaderboardEntry, error) {
	timeCondition := ""
	switch filter.TimePeriod {
	case "daily":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 day'"
	case "weekly":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 week'"
	case "monthly":
		timeCondition = "AND gs.played_at >= NOW() - INTERVAL '1 month'"
	default:
		timeCondition = ""
	}

	var orderBy string
	var selectFields string

	switch filter.Category {
	case "speed":
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
	default:
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
	args := []interface{}{gameType}
	argCount := 1

	if filter.UserID != nil {
		argCount++
		userFilter = fmt.Sprintf("AND u.id = $%d", argCount)
		args = append(args, *filter.UserID)
	}

	argCount++
	limitPlaceholder := fmt.Sprintf("$%d", argCount)
	args = append(args, limit)

	// Build the query with safe string concatenation (not user input)
	query := "SELECT " + selectFields + `
        FROM users u
        JOIN game_scores gs ON u.id = gs.user_id
        WHERE gs.game_type = $1 ` + timeCondition + ` ` + userFilter + `
        GROUP BY u.id, u.username
        ORDER BY ` + orderBy + `
        LIMIT ` + limitPlaceholder

	rows, err := db.conn.Query(query, args...)
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

func (db *DB) GetUserAchievements(userID int, gameType string) ([]string, error) {
	var achievements []string

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

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &game.AdditionalData); err != nil {
				game.AdditionalData = make(map[string]interface{})
			}
		}

		if game.AdditionalData == nil {
			game.AdditionalData = make(map[string]interface{})
		}
		game.AdditionalData["username"] = username

		games = append(games, game)
	}

	return games, nil
}

func (db *DB) CreateMultiplayerRoom(room *MultiplayerRoom) error {
	settingsJSON, err := json.Marshal(room.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO multiplayer_rooms (id, name, game_type, max_players, created_by, settings)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = db.conn.Exec(query, room.ID, room.Name, room.GameType, room.MaxPlayers, room.CreatedBy, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

func (db *DB) GetMultiplayerRoom(roomID string) (*MultiplayerRoom, error) {
	query := `
		SELECT r.id, r.name, r.game_type, r.max_players, r.status, r.created_by, 
		       r.created_at, r.started_at, r.finished_at, r.settings
		FROM multiplayer_rooms r
		WHERE r.id = $1
	`

	var room MultiplayerRoom
	var settingsJSON []byte
	err := db.conn.QueryRow(query, roomID).Scan(
		&room.ID, &room.Name, &room.GameType, &room.MaxPlayers, &room.Status,
		&room.CreatedBy, &room.CreatedAt, &room.StartedAt, &room.FinishedAt, &settingsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &room.Settings); err != nil {
			room.Settings = make(map[string]interface{})
		}
	}

	players, err := db.GetRoomPlayers(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room players: %w", err)
	}
	room.Players = players

	return &room, nil
}

func (db *DB) GetAvailableRooms(gameType string) ([]MultiplayerRoom, error) {
	query := `
		SELECT r.id, r.name, r.game_type, r.max_players, r.status, r.created_by, 
		       r.created_at, r.started_at, r.finished_at, r.settings,
		       COUNT(p.user_id) as current_players
		FROM multiplayer_rooms r
		LEFT JOIN multiplayer_players p ON r.id = p.room_id
		WHERE r.game_type = $1 AND r.status = 'waiting'
		GROUP BY r.id, r.name, r.game_type, r.max_players, r.status, r.created_by, 
		         r.created_at, r.started_at, r.finished_at, r.settings
		HAVING COUNT(p.user_id) < r.max_players
		ORDER BY r.created_at DESC
	`

	rows, err := db.conn.Query(query, gameType)
	if err != nil {
		return nil, fmt.Errorf("failed to get available rooms: %w", err)
	}
	defer rows.Close()

	var rooms []MultiplayerRoom
	for rows.Next() {
		var room MultiplayerRoom
		var settingsJSON []byte
		var currentPlayers int

		err := rows.Scan(
			&room.ID, &room.Name, &room.GameType, &room.MaxPlayers, &room.Status,
			&room.CreatedBy, &room.CreatedAt, &room.StartedAt, &room.FinishedAt,
			&settingsJSON, &currentPlayers,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		if len(settingsJSON) > 0 {
			if err := json.Unmarshal(settingsJSON, &room.Settings); err != nil {
				room.Settings = make(map[string]interface{})
			}
		}

		if room.Settings == nil {
			room.Settings = make(map[string]interface{})
		}
		room.Settings["current_players"] = currentPlayers

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (db *DB) JoinMultiplayerRoom(roomID string, userID int) error {
	room, err := db.GetMultiplayerRoom(roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	if room.Status != "waiting" {
		return fmt.Errorf("room is not accepting new players")
	}

	if len(room.Players) >= room.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	query := `
		INSERT INTO multiplayer_players (room_id, user_id, is_ready)
		VALUES ($1, $2, false)
		ON CONFLICT (room_id, user_id) DO NOTHING
	`
	_, err = db.conn.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}

	return nil
}

func (db *DB) LeaveMultiplayerRoom(roomID string, userID int) error {
	query := `DELETE FROM multiplayer_players WHERE room_id = $1 AND user_id = $2`
	_, err := db.conn.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave room: %w", err)
	}

	countQuery := `SELECT COUNT(*) FROM multiplayer_players WHERE room_id = $1`
	var playerCount int
	err = db.conn.QueryRow(countQuery, roomID).Scan(&playerCount)
	if err == nil && playerCount == 0 {
		deleteQuery := `DELETE FROM multiplayer_rooms WHERE id = $1`
		if _, err := db.conn.Exec(deleteQuery, roomID); err != nil {
			log.Printf("Error deleting empty room %s: %v", roomID, err)
		}
	}

	return nil
}

func (db *DB) GetRoomPlayers(roomID string) ([]MultiplayerPlayer, error) {
	query := `
		SELECT p.user_id, u.username, p.position, p.score, p.status, 
		       p.joined_at, p.finished_at, p.game_state, p.is_ready
		FROM multiplayer_players p
		JOIN users u ON p.user_id = u.id
		WHERE p.room_id = $1
		ORDER BY p.joined_at ASC
	`

	rows, err := db.conn.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room players: %w", err)
	}
	defer rows.Close()

	var players []MultiplayerPlayer
	for rows.Next() {
		var player MultiplayerPlayer
		var gameStateJSON []byte

		err := rows.Scan(
			&player.UserID, &player.Username, &player.Position, &player.Score,
			&player.Status, &player.JoinedAt, &player.FinishedAt, &gameStateJSON, &player.IsReady,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}

		if len(gameStateJSON) > 0 {
			if err := json.Unmarshal(gameStateJSON, &player.GameState); err != nil {
				player.GameState = make(map[string]interface{})
			}
		}

		players = append(players, player)
	}

	return players, nil
}

func (db *DB) UpdatePlayerReady(roomID string, userID int, isReady bool) error {
	query := `
		UPDATE multiplayer_players 
		SET is_ready = $3 
		WHERE room_id = $1 AND user_id = $2
	`
	_, err := db.conn.Exec(query, roomID, userID, isReady)
	if err != nil {
		return fmt.Errorf("failed to update player ready status: %w", err)
	}

	return nil
}

func (db *DB) StartMultiplayerGame(roomID string) error {
	query := `
		SELECT COUNT(*) as total, COUNT(CASE WHEN is_ready THEN 1 END) as ready
		FROM multiplayer_players 
		WHERE room_id = $1
	`
	var total, ready int
	err := db.conn.QueryRow(query, roomID).Scan(&total, &ready)
	if err != nil {
		return fmt.Errorf("failed to check player ready status: %w", err)
	}

	if total < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}

	if ready != total {
		return fmt.Errorf("not all players are ready")
	}

	updateQuery := `
		UPDATE multiplayer_rooms 
		SET status = 'playing', started_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err = db.conn.Exec(updateQuery, roomID)
	if err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

func (db *DB) UpdatePlayerGameState(roomID string, userID int, gameState map[string]interface{}, score int) error {
	gameStateJSON, err := json.Marshal(gameState)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	query := `
		UPDATE multiplayer_players 
		SET game_state = $3, score = $4 
		WHERE room_id = $1 AND user_id = $2
	`
	_, err = db.conn.Exec(query, roomID, userID, gameStateJSON, score)
	if err != nil {
		return fmt.Errorf("failed to update player game state: %w", err)
	}

	return nil
}

func (db *DB) FinishPlayerGame(roomID string, userID int, finalScore int, position int) error {
	query := `
		UPDATE multiplayer_players 
		SET status = 'finished', finished_at = CURRENT_TIMESTAMP, score = $3, position = $4
		WHERE room_id = $1 AND user_id = $2
	`
	_, err := db.conn.Exec(query, roomID, userID, finalScore, position)
	if err != nil {
		return fmt.Errorf("failed to finish player game: %w", err)
	}

	return nil
}

func (db *DB) CleanupInactiveRooms(maxAge time.Duration) ([]string, error) {
	cutoffTime := time.Now().UTC().Add(-maxAge)
	log.Printf("Running cleanup for rooms older than %v (cutoff UTC: %v)", maxAge, cutoffTime)

	selectQuery := `
		SELECT id, name, created_at FROM multiplayer_rooms 
		WHERE status = 'waiting' AND created_at < $1
	`

	rows, err := db.conn.Query(selectQuery, cutoffTime)
	if err != nil {
		log.Printf("Failed to query inactive rooms: %v", err)
		return nil, fmt.Errorf("failed to query inactive rooms: %w", err)
	}
	defer rows.Close()

	var roomsToCleanup []string
	var roomData []struct {
		ID        string
		Name      string
		CreatedAt time.Time
	}

	for rows.Next() {
		var room struct {
			ID        string
			Name      string
			CreatedAt time.Time
		}
		if err := rows.Scan(&room.ID, &room.Name, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		roomsToCleanup = append(roomsToCleanup, room.ID)
		roomData = append(roomData, room)
		log.Printf("Found room to cleanup: %s (%s) created at %v", room.Name, room.ID, room.CreatedAt)
	}

	log.Printf("Found %d rooms to cleanup: %v", len(roomsToCleanup), roomsToCleanup)

	if len(roomsToCleanup) == 0 {
		return roomsToCleanup, nil
	}

	deletePlayersQuery := `
		DELETE FROM multiplayer_players 
		WHERE room_id = ANY($1)
	`
	_, err = db.conn.Exec(deletePlayersQuery, pq.Array(roomsToCleanup))
	if err != nil {
		log.Printf("Failed to delete players from inactive rooms: %v", err)
		return nil, fmt.Errorf("failed to delete players from inactive rooms: %w", err)
	}

	deleteRoomsQuery := `
		DELETE FROM multiplayer_rooms 
		WHERE id = ANY($1)
	`
	_, err = db.conn.Exec(deleteRoomsQuery, pq.Array(roomsToCleanup))
	if err != nil {
		log.Printf("Failed to delete inactive rooms: %v", err)
		return nil, fmt.Errorf("failed to delete inactive rooms: %w", err)
	}

	for _, room := range roomData {
		log.Printf("Cleaned up inactive room: %s (%s)", room.Name, room.ID)
	}

	return roomsToCleanup, nil
}

func (db *DB) CalculatePlayerPosition(roomID string, score int) (int, error) {
	query := `
		SELECT COUNT(*) + 1 
		FROM multiplayer_players 
		WHERE room_id = $1 AND status = 'finished' AND score > $2
	`
	var position int
	err := db.conn.QueryRow(query, roomID, score).Scan(&position)
	if err != nil {
		return 1, fmt.Errorf("failed to calculate position: %w", err)
	}
	return position, nil
}

func (db *DB) GetFinishedPlayerCount(roomID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM multiplayer_players 
		WHERE room_id = $1 AND status = 'finished'
	`
	var count int
	err := db.conn.QueryRow(query, roomID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count finished players: %w", err)
	}
	return count, nil
}

func (db *DB) GetGameResults(roomID string) ([]map[string]interface{}, error) {
	query := `
		SELECT mp.user_id, u.username, mp.score, mp.position, mp.finished_at
		FROM multiplayer_players mp
		JOIN users u ON mp.user_id = u.id
		WHERE mp.room_id = $1 AND mp.status = 'finished'
		ORDER BY mp.position ASC
	`
	rows, err := db.conn.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game results: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var userID int
		var username string
		var score, position int
		var finishedAt time.Time

		err := rows.Scan(&userID, &username, &score, &position, &finishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		results = append(results, map[string]interface{}{
			"userID":     userID,
			"username":   username,
			"score":      score,
			"position":   position,
			"finishedAt": finishedAt,
		})
	}

	return results, nil
}

func (db *DB) UpdateRoomStatus(roomID string, status string) error {
	query := `UPDATE multiplayer_rooms SET status = $2 WHERE id = $1`
	_, err := db.conn.Exec(query, roomID, status)
	if err != nil {
		return fmt.Errorf("failed to update room status: %w", err)
	}
	return nil
}

func (db *DB) GetUsernameByID(userID int) (string, error) {
	var username string
	query := `SELECT username FROM users WHERE id = $1`
	err := db.conn.QueryRow(query, userID).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}
	return username, nil
}

func (db *DB) UpdatePlayerStatus(roomID string, userID int, status string) error {
	query := `
		UPDATE multiplayer_players 
		SET status = $3
		WHERE room_id = $1 AND user_id = $2
	`
	_, err := db.conn.Exec(query, roomID, userID, status)
	if err != nil {
		return fmt.Errorf("failed to update player status: %w", err)
	}
	return nil
}

func (db *DB) UpdateRoomSettings(roomID string, settings map[string]interface{}) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	query := `UPDATE multiplayer_rooms SET settings = $1 WHERE id = $2`
	_, err = db.conn.Exec(query, settingsJSON, roomID)
	if err != nil {
		return fmt.Errorf("failed to update room settings: %w", err)
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
