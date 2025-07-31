package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/isaacjstriker/devware/internal/database"
)

// generateRoomID generates a random room ID
func generateRoomID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Request structs for multiplayer endpoints
type CreateRoomRequest struct {
	Name       string                 `json:"name"`
	GameType   string                 `json:"game_type"`
	MaxPlayers int                    `json:"max_players"`
	IsPrivate  bool                   `json:"is_private"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

type JoinRoomRequest struct {
	// Empty for now, room ID comes from URL path
}

// handleCreateRoom creates a new multiplayer room
func (s *APIServer) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	log.Printf("Creating room: received request")

	// Get user from context (injected by requireAuth middleware)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		log.Printf("Creating room: user not found in context")
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "user not found in context"})
		return
	}
	log.Printf("Creating room: user found - %s (ID: %d)", user.Username, user.UserID)

	var req CreateRoomRequest
	if err := readJSON(r, &req); err != nil {
		log.Printf("Creating room: failed to parse request - %v", err)
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request format"})
		return
	}
	log.Printf("Creating room: parsed request - Name: %s, GameType: %s, MaxPlayers: %d", req.Name, req.GameType, req.MaxPlayers)

	// Create room object
	room := &database.MultiplayerRoom{
		ID:         generateRoomID(),
		Name:       req.Name,
		GameType:   req.GameType,
		MaxPlayers: req.MaxPlayers,
		Status:     "waiting",
		CreatedBy:  user.UserID,
		CreatedAt:  time.Now(),
		Settings:   req.Settings,
		Players:    []database.MultiplayerPlayer{},
		Spectators: []int{},
	}
	log.Printf("Creating room: room object created - ID: %s", room.ID)

	if err := s.db.CreateMultiplayerRoom(room); err != nil {
		log.Printf("Creating room: database error - %v", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to create room"})
		return
	}
	log.Printf("Creating room: successfully created room %s", room.ID)

	// Automatically add the creator as a player in the room
	if err := s.db.JoinMultiplayerRoom(room.ID, user.UserID); err != nil {
		log.Printf("Creating room: failed to add creator as player - %v", err)
		// Don't fail the room creation if we can't add the creator as a player
		// The user can manually join later
	} else {
		log.Printf("Creating room: successfully added creator as player")
	}

	// Get the updated room with player information
	updatedRoom, err := s.db.GetMultiplayerRoom(room.ID)
	if err != nil {
		log.Printf("Creating room: failed to get updated room - %v", err)
		// Return the original room if we can't get the updated one
		writeJSON(w, http.StatusCreated, room)
	} else {
		writeJSON(w, http.StatusCreated, updatedRoom)
	}
}

// handleJoinRoom joins a multiplayer room
func (s *APIServer) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	// Get user from context (injected by requireAuth middleware)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "user not found in context"})
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "room ID is required"})
		return
	}

	if err := s.db.JoinMultiplayerRoom(roomID, user.UserID); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	// Get updated room info
	room, err := s.db.GetMultiplayerRoom(roomID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to get room"})
		return
	}

	writeJSON(w, http.StatusOK, room)
}

// handleLeaveRoom leaves a multiplayer room
func (s *APIServer) handleLeaveRoom(w http.ResponseWriter, r *http.Request) {
	// Get user from context (injected by requireAuth middleware)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "user not found in context"})
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "room ID is required"})
		return
	}

	if err := s.db.LeaveMultiplayerRoom(roomID, user.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to leave room"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "left room successfully"})
}

// handleGetRoom gets room details
func (s *APIServer) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")

	if roomID == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "room ID is required"})
		return
	}

	room, err := s.db.GetMultiplayerRoom(roomID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, apiError{Error: "room not found"})
		return
	}

	writeJSON(w, http.StatusOK, room)
}

// handleGetAvailableRooms gets available rooms for a game type
func (s *APIServer) handleGetAvailableRooms(w http.ResponseWriter, r *http.Request) {
	gameType := r.PathValue("gameType")

	if gameType == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "game type is required"})
		return
	}

	rooms, err := s.db.GetAvailableRooms(gameType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to get rooms"})
		return
	}

	writeJSON(w, http.StatusOK, rooms)
}

// handlePlayerReady sets player ready status
func (s *APIServer) handlePlayerReady(w http.ResponseWriter, r *http.Request) {
	log.Printf("Player ready: received request")

	// Get user from context (injected by requireAuth middleware)
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		log.Printf("Player ready: user not found in context")
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "user not found in context"})
		return
	}

	roomId := r.PathValue("roomId")
	if roomId == "" {
		log.Printf("Player ready: room ID is required")
		writeJSON(w, http.StatusBadRequest, apiError{Error: "room ID is required"})
		return
	}

	log.Printf("Player ready: user %s (%d) trying to ready in room %s", user.Username, user.UserID, roomId)

	// Get current room to check player status
	room, err := s.db.GetMultiplayerRoom(roomId)
	if err != nil {
		log.Printf("Player ready: room not found: %v", err)
		writeJSON(w, http.StatusNotFound, apiError{Error: "room not found"})
		return
	}

	log.Printf("Player ready: found room with %d players", len(room.Players))
	for _, player := range room.Players {
		log.Printf("Player ready: room has player %d (ready: %v)", player.UserID, player.IsReady)
	}

	// Find current player and toggle ready status
	var currentReady bool
	found := false
	for _, player := range room.Players {
		if player.UserID == user.UserID {
			currentReady = player.IsReady
			found = true
			break
		}
	}

	if !found {
		log.Printf("Player ready: player %d not in room %s", user.UserID, roomId)
		writeJSON(w, http.StatusBadRequest, apiError{Error: "player not in room"})
		return
	}

	// Toggle ready status
	newReady := !currentReady
	log.Printf("Player ready: toggling ready status from %v to %v", currentReady, newReady)
	err = s.db.UpdatePlayerReady(roomId, user.UserID, newReady)
	if err != nil {
		log.Printf("Player ready: failed to update ready status: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to update ready status"})
		return
	}

	log.Printf("Player ready: successfully updated ready status to %v", newReady)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "updated",
		"ready":  newReady,
	})
}

// handleWebSocket handles WebSocket connections for multiplayer
func (s *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	s.wsHub.ServeWS(w, r)
}
