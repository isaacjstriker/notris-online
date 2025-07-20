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
		if db.dbType == "postgres" {
			err := db.conn.QueryRow(
				"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
				user.username, user.email, user.password,
			).Scan(&userID)
			if err != nil {
				return fmt.Errorf("failed to create test user %s: %w", user.username, err)
			}
		} else {
			result, err := db.conn.Exec(
				"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
				user.username, user.email, user.password,
			)
			if err != nil {
				return fmt.Errorf("failed to create test user %s: %w", user.username, err)
			}
			id, _ := result.LastInsertId()
			userID = int(id)
		}
		userIDs[i] = userID
		fmt.Printf("   ✅ Created user: %s (ID: %d)\n", user.username, userID)
	}

	// Create test scores for typing game
	testScores := []struct {
		userID int
		scores []struct {
			score    int
			metadata map[string]interface{}
		}
	}{
		{userIDs[0], []struct {
			score    int
			metadata map[string]interface{}
		}{
			{150, map[string]interface{}{"wpm": 45, "accuracy": 92.5}},
			{175, map[string]interface{}{"wpm": 52, "accuracy": 94.0}},
			{200, map[string]interface{}{"wpm": 58, "accuracy": 96.2}},
			{180, map[string]interface{}{"wpm": 54, "accuracy": 93.8}},
			{165, map[string]interface{}{"wpm": 48, "accuracy": 95.1}},
		}},
		{userIDs[1], []struct {
			score    int
			metadata map[string]interface{}
		}{
			{220, map[string]interface{}{"wpm": 68, "accuracy": 97.2}},
			{250, map[string]interface{}{"wpm": 74, "accuracy": 98.1}},
			{275, map[string]interface{}{"wpm": 81, "accuracy": 98.8}},
			{240, map[string]interface{}{"wpm": 72, "accuracy": 97.5}},
			{260, map[string]interface{}{"wpm": 76, "accuracy": 98.3}},
		}},
		{userIDs[2], []struct {
			score    int
			metadata map[string]interface{}
		}{
			{180, map[string]interface{}{"wpm": 55, "accuracy": 94.2}},
			{190, map[string]interface{}{"wpm": 58, "accuracy": 95.0}},
			{185, map[string]interface{}{"wpm": 56, "accuracy": 94.8}},
			{200, map[string]interface{}{"wpm": 61, "accuracy": 96.1}},
			{195, map[string]interface{}{"wpm": 59, "accuracy": 95.5}},
		}},
		{userIDs[3], []struct {
			score    int
			metadata map[string]interface{}
		}{
			{310, map[string]interface{}{"wpm": 92, "accuracy": 99.2}},
			{295, map[string]interface{}{"wpm": 88, "accuracy": 98.9}},
			{320, map[string]interface{}{"wpm": 95, "accuracy": 99.5}},
			{305, map[string]interface{}{"wpm": 91, "accuracy": 99.1}},
		}},
		{userIDs[4], []struct {
			score    int
			metadata map[string]interface{}
		}{
			{140, map[string]interface{}{"wpm": 42, "accuracy": 91.8}},
			{155, map[string]interface{}{"wpm": 46, "accuracy": 93.2}},
			{160, map[string]interface{}{"wpm": 48, "accuracy": 93.8}},
		}},
	}

	for _, userScore := range testScores {
		for j, scoreData := range userScore.scores {
			var metadataValue interface{}

			if db.dbType == "postgres" {
				// PostgreSQL uses JSONB
				metadataJSON, err := json.Marshal(scoreData.metadata)
				if err != nil {
					return fmt.Errorf("failed to marshal metadata: %w", err)
				}
				metadataValue = metadataJSON

				_, err = db.conn.Exec(
					"INSERT INTO game_scores (user_id, game_type, score, metadata, played_at) VALUES ($1, $2, $3, $4, $5)",
					userScore.userID, "typing", scoreData.score, metadataValue,
					time.Now().Add(-time.Duration(len(userScore.scores)-j)*24*time.Hour),
				)
			} else {
				// SQLite uses TEXT for JSON
				metadataJSON, err := json.Marshal(scoreData.metadata)
				if err != nil {
					return fmt.Errorf("failed to marshal metadata: %w", err)
				}
				metadataValue = string(metadataJSON)

				_, err = db.conn.Exec(
					"INSERT INTO game_scores (user_id, game_type, score, metadata, played_at) VALUES (?, ?, ?, ?, ?)",
					userScore.userID, "typing", scoreData.score, metadataValue,
					time.Now().Add(-time.Duration(len(userScore.scores)-j)*24*time.Hour),
				)
			}
		}
		fmt.Printf("   📊 Created %d scores for user ID %d\n", len(userScore.scores), userScore.userID)
	}

	fmt.Println("✅ Sample data created successfully!")
	return nil
}

// SubmitScore submits a new score to the database
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
