package api

import (
	"net/http"
	"strconv"

	"github.com/isaacjstriker/devware/internal/database"
)

// handleGetLeaderboard handles requests for game leaderboards
func (s *APIServer) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	gameType := r.PathValue("gameType")
	if gameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "gameType is required"})
		return
	}

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 15
	}

	timePeriod := r.URL.Query().Get("period")
	if timePeriod == "" {
		timePeriod = "all"
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = "score"
	}

	// Create filter
	filter := database.LeaderboardFilter{
		TimePeriod: timePeriod,
		Category:   category,
	}

	// Get filtered leaderboard
	entries, err := s.db.GetFilteredLeaderboard(gameType, limit, filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to fetch leaderboard"})
		return
	}

	// Add achievements for each user if requested
	if r.URL.Query().Get("include_achievements") == "true" {
		for i := range entries {
			userID, err := s.getUserIDByUsername(entries[i].Username)
			if err == nil {
				achievements, err := s.db.GetUserAchievements(userID, gameType)
				if err == nil {
					entries[i].Achievements = achievements
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, entries)
}

// handleGetRecentGames handles requests for recent game activity
func (s *APIServer) handleGetRecentGames(w http.ResponseWriter, r *http.Request) {
	gameType := r.PathValue("gameType")
	if gameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "gameType is required"})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	games, err := s.db.GetRecentGames(gameType, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to fetch recent games"})
		return
	}

	writeJSON(w, http.StatusOK, games)
}

// Helper function to get user ID by username
func (s *APIServer) getUserIDByUsername(username string) (int, error) {
	user, _, err := s.db.GetUserByUsername(username)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}
