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

func (s *APIServer) handleSubmitScore(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "authorization header required"})
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid authorization format"})
		return
	}
	tokenString := authHeader[len(bearerPrefix):]

	userInfo, err := s.validateJWT(tokenString)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid token"})
		return
	}

	var submission ScoreSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	if submission.GameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "game_type is required"})
		return
	}

	if submission.Score < 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "score must be non-negative"})
		return
	}

	err = s.db.SaveGameScore(userInfo.UserID, submission.GameType, submission.Score, submission.Metadata)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to save score"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Score saved successfully",
	})
}
