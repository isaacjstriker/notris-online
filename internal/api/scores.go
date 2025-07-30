package api

import (
	"encoding/json"
	"net/http"
)

type ScoreSubmission struct {
	GameType string                 `json:"game_type"`
	Score    int                    `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// handleSubmitScore handles score submission from authenticated users
func (s *APIServer) handleSubmitScore(w http.ResponseWriter, r *http.Request) {
	// Get JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "authorization header required"})
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid authorization format"})
		return
	}
	tokenString := authHeader[len(bearerPrefix):]

	// Validate JWT token and get user info
	userInfo, err := s.validateJWT(tokenString)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid token"})
		return
	}

	// Parse score submission
	var submission ScoreSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	// Validate required fields
	if submission.GameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "game_type is required"})
		return
	}

	if submission.Score < 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "score must be non-negative"})
		return
	}

	// Save score to database
	err = s.db.SaveGameScore(userInfo.UserID, submission.GameType, submission.Score, submission.Metadata)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to save score"})
		return
	}

	// Return success response
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Score saved successfully",
	})
}
