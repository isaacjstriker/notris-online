package api

import (
	"net/http"
	"strconv"
)

// handleGetLeaderboard handles requests for game leaderboards
func (s *APIServer) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	gameType := r.PathValue("gameType")
	if gameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "gameType is required"})
		return
	}

	// Get limit from query params, default to 15
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 15
	}

	entries, err := s.db.GetLeaderboard(gameType, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to fetch leaderboard"})
		return
	}

	writeJSON(w, http.StatusOK, entries)
}
