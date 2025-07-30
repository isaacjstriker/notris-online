package database

import (
	"encoding/json"
	"fmt"
	"time"
)

// CreateTestData creates sample data for testing and development
func (db *DB) CreateTestData() error {
	// Only create test data if no users exist
	var userCount int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}

	if userCount > 0 {
		return nil // Data already exists
	}

	fmt.Println("🎮 Creating sample data for testing...")

	// Create test users
	testUsers := []struct {
		username string
		email    string
		password string
	}{
		{"speedster", "speedster@example.com", "hashed_password_123"},
		{"typingpro", "pro@example.com", "hashed_password_456"},
		{"quickfingers", "quick@example.com", "hashed_password_789"},
		{"gamemaster", "master@example.com", "hashed_password_abc"},
		{"challenger", "challenger@example.com", "hashed_password_def"},
	}

	userIDs := make([]int, len(testUsers))

	for i, user := range testUsers {
		var userID int
		err := db.conn.QueryRow(
			"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
			user.username, user.email, user.password,
		).Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to create test user %s: %w", user.username, err)
		}
		userIDs[i] = userID
		fmt.Printf("   ✅ Created user: %s (ID: %d)\n", user.username, userID)
	}

	// Create test scores for both typing and tetris games
	testScores := []struct {
		userID int
		scores []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}
	}{
		{userIDs[0], []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}{
			{"typing", 150, map[string]interface{}{"wpm": 45, "accuracy": 92.5}},
			{"typing", 175, map[string]interface{}{"wpm": 52, "accuracy": 94.0}},
			{"tetris", 8500, map[string]interface{}{"lines": 25, "level": 3, "game_time": 120.5}},
			{"tetris", 12000, map[string]interface{}{"lines": 35, "level": 4, "game_time": 180.2}},
		}},
		{userIDs[1], []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}{
			{"typing", 220, map[string]interface{}{"wpm": 68, "accuracy": 97.2}},
			{"tetris", 25000, map[string]interface{}{"lines": 60, "level": 7, "game_time": 300.8}},
			{"tetris", 18500, map[string]interface{}{"lines": 45, "level": 5, "game_time": 210.1}},
		}},
		{userIDs[2], []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}{
			{"typing", 180, map[string]interface{}{"wpm": 55, "accuracy": 94.2}},
			{"typing", 190, map[string]interface{}{"wpm": 58, "accuracy": 95.0}},
			{"tetris", 9000, map[string]interface{}{"lines": 30, "level": 4, "game_time": 150.7}},
			{"tetris", 13000, map[string]interface{}{"lines": 40, "level": 5, "game_time": 200.3}},
		}},
		{userIDs[3], []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}{
			{"typing", 310, map[string]interface{}{"wpm": 92, "accuracy": 99.2}},
			{"tetris", 27000, map[string]interface{}{"lines": 70, "level": 8, "game_time": 360.9}},
			{"tetris", 22000, map[string]interface{}{"lines": 55, "level": 6, "game_time": 250.4}},
		}},
		{userIDs[4], []struct {
			gameType string
			score    int
			metadata map[string]interface{}
		}{
			{"typing", 140, map[string]interface{}{"wpm": 42, "accuracy": 91.8}},
			{"tetris", 7000, map[string]interface{}{"lines": 20, "level": 2, "game_time": 100.1}},
			{"tetris", 9500, map[string]interface{}{"lines": 28, "level": 3, "game_time": 140.6}},
		}},
	}

	for _, userScore := range testScores {
		for j, scoreData := range userScore.scores {
			// PostgreSQL uses JSONB
			metadataJSON, err := json.Marshal(scoreData.metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}

			_, err = db.conn.Exec(
				"INSERT INTO game_scores (user_id, game_type, score, metadata, played_at) VALUES ($1, $2, $3, $4, $5)",
				userScore.userID, scoreData.gameType, scoreData.score, metadataJSON,
				time.Now().Add(-time.Duration(len(userScore.scores)-j)*12*time.Hour),
			)
			if err != nil {
				return fmt.Errorf("failed to create test score: %w", err)
			}
		}
		fmt.Printf("   📊 Created %d scores for user ID %d\n", len(userScore.scores), userScore.userID)
	}

	fmt.Println("✅ Sample data created successfully!")
	return nil
}
