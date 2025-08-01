package multiplayer

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/isaacjstriker/devware/games/tetris"
	"github.com/isaacjstriker/devware/internal/database"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type   string                 `json:"type"`
	RoomID string                 `json:"room_id,omitempty"`
	UserID int                    `json:"user_id,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	UserID int
	RoomID string
	Conn   *websocket.Conn
	Send   chan WebSocketMessage
	Hub    *Hub
}

// UserInfo represents user information from JWT token
type UserInfo struct {
	ID       int    `json:"user_id"`
	Username string `json:"username"`
}

// JWTValidator is a function type for validating JWT tokens
type JWTValidator func(tokenString string) (*UserInfo, error)

// MultiplayerGame represents an active multiplayer game session
type MultiplayerGame struct {
	RoomID     string
	Players    map[int]*tetris.Tetris // UserID -> Game instance
	StartTime  time.Time
	IsActive   bool
	GameTicker *time.Ticker
	mutex      sync.RWMutex
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients          map[*Client]bool
	rooms            map[string]map[*Client]bool
	multiplayerGames map[string]*MultiplayerGame // RoomID -> Game
	broadcast        chan WebSocketMessage
	register         chan *Client
	unregister       chan *Client
	db               *database.DB
	mutex            sync.RWMutex
	stopCleanup      chan bool
	validateJWT      JWTValidator
}

// NewHub creates a new WebSocket hub
func NewHub(db *database.DB, jwtValidator JWTValidator) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		rooms:            make(map[string]map[*Client]bool),
		multiplayerGames: make(map[string]*MultiplayerGame),
		broadcast:        make(chan WebSocketMessage),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		db:               db,
		stopCleanup:      make(chan bool),
		validateJWT:      jwtValidator,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Start the room cleanup goroutine
	go h.startRoomCleanup()

	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			if client.RoomID != "" {
				if h.rooms[client.RoomID] == nil {
					h.rooms[client.RoomID] = make(map[*Client]bool)
				}
				h.rooms[client.RoomID][client] = true

				// Check if this is a reconnection during an active game
				go h.checkReconnection(client.UserID, client.RoomID)
			}
			h.mutex.Unlock()

			// Send welcome message
			select {
			case client.Send <- WebSocketMessage{
				Type: "connected",
				Data: map[string]interface{}{
					"user_id": client.UserID,
					"room_id": client.RoomID,
				},
			}:
			default:
				close(client.Send)
				delete(h.clients, client)
			}

			log.Printf("Client %s connected to room %s", client.ID, client.RoomID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				roomID := client.RoomID
				if roomID != "" && h.rooms[roomID] != nil {
					delete(h.rooms[roomID], client)
					if len(h.rooms[roomID]) == 0 {
						delete(h.rooms, roomID)
					}

					// Handle player disconnection during game
					go h.handlePlayerDisconnection(client.UserID, roomID)
				}
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("Client %s disconnected from room %s", client.ID, client.RoomID)

		case message := <-h.broadcast:
			h.handleMessage(message)
		}
	}
}

// startRoomCleanup runs a background task to clean up inactive rooms
func (h *Hub) startRoomCleanup() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.cleanupInactiveRooms()
		case <-h.stopCleanup:
			log.Println("Room cleanup stopped")
			return
		}
	}
}

// cleanupInactiveRooms removes rooms that have been waiting for more than 5 minutes
func (h *Hub) cleanupInactiveRooms() {
	log.Printf("Running room cleanup check...")
	maxAge := 5 * time.Minute
	cleanedRoomIDs, err := h.db.CleanupInactiveRooms(maxAge)
	if err != nil {
		log.Printf("Failed to cleanup inactive rooms: %v", err)
		return
	}

	log.Printf("Cleanup completed. Cleaned %d rooms", len(cleanedRoomIDs))

	if len(cleanedRoomIDs) > 0 {
		log.Printf("Cleaned up %d inactive rooms", len(cleanedRoomIDs))

		// Notify all clients that these rooms have been removed
		for _, roomID := range cleanedRoomIDs {
			// Disconnect any clients still connected to these rooms
			h.mutex.RLock()
			roomClients := h.rooms[roomID]
			h.mutex.RUnlock()

			for client := range roomClients {
				select {
				case client.Send <- WebSocketMessage{
					Type:   "room_closed",
					RoomID: roomID,
					Data: map[string]interface{}{
						"reason": "Room closed due to inactivity",
					},
				}:
				default:
					// Client send channel is full, close it
					close(client.Send)
					h.mutex.Lock()
					delete(h.clients, client)
					delete(h.rooms[roomID], client)
					h.mutex.Unlock()
				}
			}

			// Clean up the room from memory
			h.mutex.Lock()
			delete(h.rooms, roomID)
			h.mutex.Unlock()
		}

		// Broadcast to all clients to refresh their room lists
		h.broadcastToAll(WebSocketMessage{
			Type: "rooms_updated",
			Data: map[string]interface{}{
				"removed_rooms": cleanedRoomIDs,
				"reason":        "inactive_cleanup",
			},
		})
	}
}

// broadcastToAll sends a message to all connected clients
func (h *Hub) broadcastToAll(message WebSocketMessage) {
	h.mutex.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			h.mutex.Lock()
			delete(h.clients, client)
			if client.RoomID != "" && h.rooms[client.RoomID] != nil {
				delete(h.rooms[client.RoomID], client)
			}
			h.mutex.Unlock()
		}
	}
}

// Stop gracefully stops the hub
func (h *Hub) Stop() {
	close(h.stopCleanup)
}

// handleMessage processes different types of WebSocket messages
func (h *Hub) handleMessage(message WebSocketMessage) {
	switch message.Type {
	case "room_update":
		h.broadcastToRoom(message.RoomID, message)
	case "game_state":
		h.handleGameState(message)
	case "player_ready":
		h.handlePlayerReady(message)
	case "start_game":
		h.handleStartGame(message)
	case "start_multiplayer_game":
		h.handleStartMultiplayerGame(message)
	case "game_input":
		h.handleGameInput(message)
	case "player_finished":
		h.handlePlayerFinished(message)
	case "spectate_request":
		h.handleSpectateRequest(message)
	case "multiplayerInit":
		h.handleMultiplayerInit(message)
	case "setLevel":
		h.handleSetLevel(message)
	case "player_disconnect":
		h.handlePlayerDisconnectMessage(message)
	case "heartbeat":
		// Handle heartbeat - no action needed, just confirms connection
		log.Printf("Heartbeat received from user %d in room %s", message.UserID, message.RoomID)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}

}

// handleSetLevel broadcasts a level change to all clients in the room
func (h *Hub) handleSetLevel(message WebSocketMessage) {
	if message.RoomID == "" {
		return
	}
	// Broadcast the setLevel event to all clients
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "setLevel",
		RoomID: message.RoomID,
		Data:   message.Data,
	})
}

// handleMultiplayerInit sets the starting level for the room if provided
func (h *Hub) handleMultiplayerInit(message WebSocketMessage) {
	if message.RoomID == "" {
		return
	}
	startingLevel := 1
	if message.Data != nil {
		if lvl, ok := message.Data["startingLevel"]; ok {
			switch v := lvl.(type) {
			case float64:
				startingLevel = int(v)
			case int:
				startingLevel = v
			}
		}
	}
	// Update the room's settings in the database
	room, err := h.db.GetMultiplayerRoom(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room for multiplayerInit: %v", err)
		return
	}
	if room.Settings == nil {
		room.Settings = make(map[string]interface{})
	}
	room.Settings["starting_level"] = startingLevel
	if err := h.db.UpdateRoomSettings(message.RoomID, room.Settings); err != nil {
		log.Printf("Failed to update room settings for starting level: %v", err)
	} else {
		log.Printf("Set starting level for room %s to %d", message.RoomID, startingLevel)
	}
}

// handleGameState updates player game state and broadcasts to room
func (h *Hub) handleGameState(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		return
	}

	// Update database
	score := 0
	if s, ok := message.Data["score"].(float64); ok {
		score = int(s)
	}

	err := h.db.UpdatePlayerGameState(message.RoomID, message.UserID, message.Data, score)
	if err != nil {
		log.Printf("Failed to update game state: %v", err)
		return
	}

	// Broadcast to room
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "player_update",
		RoomID: message.RoomID,
		UserID: message.UserID,
		Data:   message.Data,
	})
}

// handleStartMultiplayerGame starts actual gameplay for all players in the room
func (h *Hub) handleStartMultiplayerGame(message WebSocketMessage) {
	if message.RoomID == "" {
		log.Printf("No room ID provided for start_multiplayer_game")
		return
	}

	log.Printf("Starting multiplayer game for room: %s", message.RoomID)

	// Get room details to determine starting level
	room, err := h.db.GetMultiplayerRoom(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room for game start: %v", err)
		return
	}

	startingLevel := 1
	if room.Settings != nil {
		if level, ok := room.Settings["starting_level"]; ok {
			if lvl, ok := level.(float64); ok {
				startingLevel = int(lvl)
			}
		}
	}

	// Create multiplayer game session
	h.mutex.Lock()
	multiplayerGame := &MultiplayerGame{
		RoomID:    message.RoomID,
		Players:   make(map[int]*tetris.Tetris),
		StartTime: time.Now(),
		IsActive:  true,
	}

	// Create individual Tetris instances for each player in the room
	for _, player := range room.Players {
		tetrisGame := tetris.NewTetris()
		tetrisGame.SetLevel(startingLevel)
		multiplayerGame.Players[player.UserID] = tetrisGame
		log.Printf("Created Tetris instance for player %d (%s) with starting level %d",
			player.UserID, player.Username, startingLevel)
	}

	h.multiplayerGames[message.RoomID] = multiplayerGame
	h.mutex.Unlock()

	// Update room status to active in database
	err = h.db.UpdateRoomStatus(message.RoomID, "active")
	if err != nil {
		log.Printf("Failed to update room status to active: %v", err)
	} else {
		log.Printf("Room %s status updated to active", message.RoomID)
	}

	// Start the game tick for this multiplayer session
	h.startMultiplayerGameTick(message.RoomID)

	// Broadcast game start message to all players in the room
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "multiplayer_game_started",
		RoomID: message.RoomID,
		Data: map[string]interface{}{
			"starting_level": startingLevel,
			"message":        "Game starting! Use arrow keys to play.",
		},
	})

	log.Printf("Multiplayer game started for room %s with %d players", message.RoomID, len(multiplayerGame.Players))
}

// startMultiplayerGameTick starts the game loop for a multiplayer game
func (h *Hub) startMultiplayerGameTick(roomID string) {
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS game loop
		defer ticker.Stop()

		for range ticker.C {
			h.mutex.RLock()
			multiplayerGame, exists := h.multiplayerGames[roomID]
			h.mutex.RUnlock()

			if !exists || !multiplayerGame.IsActive {
				log.Printf("Stopping game tick for room %s (game ended or removed)", roomID)
				return
			}

			multiplayerGame.mutex.Lock()
			// Update each player's game and broadcast their state
			for userID, tetrisGame := range multiplayerGame.Players {
				if !tetrisGame.IsGameOver() {
					tetrisGame.Update()

					// Broadcast this player's game state to all players in the room
					gameState := tetrisGame.GetState()
					h.broadcastToRoom(roomID, WebSocketMessage{
						Type:   "player_game_state",
						RoomID: roomID,
						UserID: userID,
						Data: map[string]interface{}{
							"board":      gameState.Board,
							"score":      gameState.Score,
							"level":      gameState.Level,
							"lines":      gameState.Lines,
							"gameOver":   gameState.GameOver,
							"paused":     gameState.Paused,
							"nextPiece":  gameState.NextPiece,
							"holdPiece":  gameState.HoldPiece,
							"ghostPiece": gameState.GhostPiece,
							"userID":     userID,
						},
					})
				}
			}
			multiplayerGame.mutex.Unlock()

			// Check if any player finished and handle game completion
			h.checkMultiplayerGameCompletion(roomID)
		}
	}()
}

// checkMultiplayerGameCompletion checks if the multiplayer game should end
func (h *Hub) checkMultiplayerGameCompletion(roomID string) {
	h.mutex.RLock()
	multiplayerGame, exists := h.multiplayerGames[roomID]
	h.mutex.RUnlock()

	if !exists || !multiplayerGame.IsActive {
		return
	}

	multiplayerGame.mutex.RLock()
	finishedPlayers := 0
	totalPlayers := len(multiplayerGame.Players)

	for _, tetrisGame := range multiplayerGame.Players {
		if tetrisGame.IsGameOver() {
			finishedPlayers++
		}
	}
	multiplayerGame.mutex.RUnlock()

	// End the game when at least one player finishes (traditional Tetris multiplayer)
	if finishedPlayers > 0 {
		log.Printf("Ending multiplayer game in room %s (%d/%d players finished)",
			roomID, finishedPlayers, totalPlayers)
		h.endMultiplayerGame(roomID)
	}
}

// endMultiplayerGame ends a multiplayer game session
func (h *Hub) endMultiplayerGame(roomID string) {
	h.mutex.Lock()
	multiplayerGame, exists := h.multiplayerGames[roomID]
	if !exists {
		h.mutex.Unlock()
		return
	}

	multiplayerGame.IsActive = false
	if multiplayerGame.GameTicker != nil {
		multiplayerGame.GameTicker.Stop()
	}
	delete(h.multiplayerGames, roomID)
	h.mutex.Unlock()

	// Update room status back to waiting
	err := h.db.UpdateRoomStatus(roomID, "waiting")
	if err != nil {
		log.Printf("Failed to update room status to waiting: %v", err)
	}

	// Send final results to all players
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "multiplayer_game_ended",
		RoomID: roomID,
		Data: map[string]interface{}{
			"message": "Game ended! Thank you for playing.",
		},
	})

	log.Printf("Multiplayer game ended for room %s", roomID)
}

// handleGameInput processes game input from players and broadcasts state changes
func (h *Hub) handleGameInput(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		log.Printf("Invalid game input: missing room ID or user ID")
		return
	}

	// Get the action from the message data
	action, ok := message.Data["action"].(string)
	if !ok {
		log.Printf("Invalid game input: no action specified")
		return
	}

	log.Printf("Game input from user %d in room %s: %s", message.UserID, message.RoomID, action)

	// Get the multiplayer game instance
	h.mutex.RLock()
	multiplayerGame, exists := h.multiplayerGames[message.RoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("No active multiplayer game found for room %s", message.RoomID)
		return
	}

	// Get the player's game instance
	multiplayerGame.mutex.Lock()
	tetrisGame, playerExists := multiplayerGame.Players[message.UserID]
	if !playerExists {
		multiplayerGame.mutex.Unlock()
		log.Printf("Player %d not found in game for room %s", message.UserID, message.RoomID)
		return
	}

	// Process the input through the Tetris game engine
	if !tetrisGame.IsGameOver() {
		tetrisGame.HandleWebInput(action)

		// Get the updated game state
		gameState := tetrisGame.GetState()

		// Broadcast this player's updated game state to all players in the room
		h.broadcastToRoom(message.RoomID, WebSocketMessage{
			Type:   "player_game_state",
			RoomID: message.RoomID,
			UserID: message.UserID,
			Data: map[string]interface{}{
				"board":      gameState.Board,
				"score":      gameState.Score,
				"level":      gameState.Level,
				"lines":      gameState.Lines,
				"gameOver":   gameState.GameOver,
				"paused":     gameState.Paused,
				"nextPiece":  gameState.NextPiece,
				"holdPiece":  gameState.HoldPiece,
				"ghostPiece": gameState.GhostPiece,
				"userID":     message.UserID,
			},
		})

		log.Printf("Processed input '%s' for player %d, new score: %d",
			action, message.UserID, gameState.Score)
	}
	multiplayerGame.mutex.Unlock()
}

// handlePlayerReady sets player ready status
func (h *Hub) handlePlayerReady(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		return
	}

	isReady := false
	if r, ok := message.Data["ready"].(bool); ok {
		isReady = r
	}

	err := h.db.UpdatePlayerReady(message.RoomID, message.UserID, isReady)
	if err != nil {
		log.Printf("Failed to update player ready: %v", err)
		return
	}

	// Get updated room info
	room, err := h.db.GetMultiplayerRoom(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room: %v", err)
		return
	}

	// Broadcast room update
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "room_update",
		RoomID: message.RoomID,
		Data: map[string]interface{}{
			"room": room,
		},
	})

	// Check if all players are ready and start game automatically
	h.checkAndStartGame(room)
}

// checkAndStartGame starts the game if all players are ready
func (h *Hub) checkAndStartGame(room *database.MultiplayerRoom) {
	if room.Status != "waiting" {
		return // Game already started or finished
	}

	// Count ready players
	readyCount := 0
	totalPlayers := len(room.Players)

	for _, player := range room.Players {
		if player.IsReady {
			readyCount++
		}
	}

	log.Printf("Room %s ready check: %d/%d players ready", room.ID, readyCount, totalPlayers)

	// Start game if we have at least 2 players and all are ready
	if totalPlayers >= 2 && readyCount == totalPlayers {
		log.Printf("Auto-starting game in room %s: %d/%d players ready", room.ID, readyCount, totalPlayers)

		err := h.db.StartMultiplayerGame(room.ID)
		if err != nil {
			log.Printf("Failed to auto-start game: %v", err)
			return
		}

		// Broadcast game start
		h.broadcastToRoom(room.ID, WebSocketMessage{
			Type:   "game_start",
			RoomID: room.ID,
			Data: map[string]interface{}{
				"timestamp": time.Now(),
				"message":   "All players ready! Game starting...",
			},
		})
	}
}

// handleStartGame attempts to start the game
func (h *Hub) handleStartGame(message WebSocketMessage) {
	if message.RoomID == "" {
		return
	}

	err := h.db.StartMultiplayerGame(message.RoomID)
	if err != nil {
		log.Printf("Failed to start game: %v", err)
		// Send error to requesting client
		h.broadcastToRoom(message.RoomID, WebSocketMessage{
			Type:  "error",
			Error: err.Error(),
		})
		return
	}

	// Broadcast game start
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "game_started",
		RoomID: message.RoomID,
		Data: map[string]interface{}{
			"timestamp": time.Now(),
		},
	})
}

// handlePlayerFinished marks player as finished and handles game completion
func (h *Hub) handlePlayerFinished(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		return
	}

	score := 0
	if s, ok := message.Data["score"].(float64); ok {
		score = int(s)
	}

	lines := 0
	if l, ok := message.Data["lines"].(float64); ok {
		lines = int(l)
	}

	// Calculate position based on existing finished players
	position, err := h.calculatePlayerPosition(message.RoomID, score)
	if err != nil {
		log.Printf("Failed to calculate position: %v", err)
		position = 1 // Default position
	}

	err = h.db.FinishPlayerGame(message.RoomID, message.UserID, score, position)
	if err != nil {
		log.Printf("Failed to finish player game: %v", err)
		return
	}

	// Get player username for notification
	username, err := h.getUsernameByID(message.UserID)
	if err != nil {
		log.Printf("Failed to get username: %v", err)
		username = "Unknown Player"
	}

	// Broadcast player finished to all players in room
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "player_finished",
		RoomID: message.RoomID,
		UserID: message.UserID,
		Data: map[string]interface{}{
			"playerName": username,
			"score":      score,
			"lines":      lines,
			"position":   position,
		},
	})

	// Check if game is complete and handle final results
	go h.checkGameCompletion(message.RoomID)
}

// calculatePlayerPosition determines finishing position based on score
func (h *Hub) calculatePlayerPosition(roomID string, score int) (int, error) {
	return h.db.CalculatePlayerPosition(roomID, score)
}

// checkGameCompletion checks if any player has finished and ends the game immediately
func (h *Hub) checkGameCompletion(roomID string) {
	// Count finished players
	finishedCount, err := h.db.GetFinishedPlayerCount(roomID)
	if err != nil {
		log.Printf("Failed to count finished players: %v", err)
		return
	}

	// In multiplayer Tetris, when one player finishes (gets game over), the match ends for everyone
	if finishedCount >= 1 {
		log.Printf("Player finished in room %s, ending match for all players", roomID)

		// Mark remaining active players as finished with their current scores
		h.finishRemainingPlayers(roomID)

		// Send final results
		h.sendFinalResults(roomID)
	}
}

// finishRemainingPlayers marks all remaining active players as finished
func (h *Hub) finishRemainingPlayers(roomID string) {
	// Get all players in the room
	players, err := h.db.GetRoomPlayers(roomID)
	if err != nil {
		log.Printf("Failed to get room players for finishing: %v", err)
		return
	}

	for _, player := range players {
		// If player hasn't finished yet, mark them as finished
		if player.Status != "finished" {
			// Calculate their position (they get last place since they didn't finish naturally)
			position, err := h.db.CalculatePlayerPosition(roomID, player.Score)
			if err != nil {
				position = len(players) // Default to last position
			}

			err = h.db.FinishPlayerGame(roomID, player.UserID, player.Score, position)
			if err != nil {
				log.Printf("Failed to finish remaining player %d: %v", player.UserID, err)
			} else {
				log.Printf("Marked player %s as finished due to match end", player.Username)
			}
		}
	}
}

// sendFinalResults broadcasts final game results to all players
func (h *Hub) sendFinalResults(roomID string) {
	// Get final standings
	results, err := h.db.GetGameResults(roomID)
	if err != nil {
		log.Printf("Failed to get final results: %v", err)
		return
	}

	// Broadcast final results
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "game_complete",
		RoomID: roomID,
		Data: map[string]interface{}{
			"results": results,
		},
	})

	// Update room status to completed
	err = h.db.UpdateRoomStatus(roomID, "completed")
	if err != nil {
		log.Printf("Failed to update room status: %v", err)
	}

	log.Printf("Game completed for room %s with %d players", roomID, len(results))
}

// getUsernameByID gets username by user ID
func (h *Hub) getUsernameByID(userID int) (string, error) {
	return h.db.GetUsernameByID(userID)
}

// handlePlayerDisconnection handles when a player disconnects during a game
func (h *Hub) handlePlayerDisconnection(userID int, roomID string) {
	// Get room info to check game status
	room, err := h.db.GetMultiplayerRoom(roomID)
	if err != nil {
		log.Printf("Failed to get room for disconnection handling: %v", err)
		return
	}

	// Only handle disconnections during active games
	if room.Status != "active" {
		return
	}

	// Get player info
	username, err := h.getUsernameByID(userID)
	if err != nil {
		log.Printf("Failed to get username for disconnected player: %v", err)
		username = "Unknown Player"
	}

	// Mark player as disconnected in database
	err = h.db.UpdatePlayerStatus(roomID, userID, "disconnected")
	if err != nil {
		log.Printf("Failed to update player status to disconnected: %v", err)
	}

	// Notify other players about disconnection
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "player_disconnected",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"playerName": username,
			"message":    fmt.Sprintf("%s has disconnected", username),
		},
	})

	// Check if we should pause the game (if there are other players still connected)
	h.mutex.RLock()
	activeClients := len(h.rooms[roomID])
	h.mutex.RUnlock()

	// Check if this is a multiplayer game session
	h.mutex.RLock()
	multiplayerGame, isMultiplayerGame := h.multiplayerGames[roomID]
	h.mutex.RUnlock()

	if isMultiplayerGame && multiplayerGame.IsActive {
		// For multiplayer games, immediately end the game when a player disconnects
		log.Printf("Player %s disconnected from active multiplayer game in room %s, ending game", username, roomID)

		// End the multiplayer game immediately
		h.endMultiplayerGame(roomID)

		// Notify remaining players that the match ended due to disconnection
		h.broadcastToRoom(roomID, WebSocketMessage{
			Type:   "match_ended",
			RoomID: roomID,
			Data: map[string]interface{}{
				"reason":     "player_disconnected",
				"message":    fmt.Sprintf("Match ended because %s disconnected", username),
				"playerName": username,
			},
		})
	} else if activeClients > 0 {
		// For non-multiplayer games, start disconnection timer (30 seconds to reconnect)
		go h.startDisconnectionTimer(userID, roomID, username, 30*time.Second)
	}
}

// startDisconnectionTimer gives a player time to reconnect before ending their game
func (h *Hub) startDisconnectionTimer(userID int, roomID string, username string, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	<-timer.C

	// Check if player has reconnected
	h.mutex.RLock()
	playerReconnected := false
	if roomClients, exists := h.rooms[roomID]; exists {
		for client := range roomClients {
			if client.UserID == userID {
				playerReconnected = true
				break
			}
		}
	}
	h.mutex.RUnlock()

	if !playerReconnected {
		// Player didn't reconnect in time, end their game
		log.Printf("Player %s didn't reconnect to room %s in time, ending their game", username, roomID)

		// Mark as finished with 0 score due to disconnection
		err := h.db.FinishPlayerGame(roomID, userID, 0, 999) // Use high position number for disconnected
		if err != nil {
			log.Printf("Failed to finish disconnected player game: %v", err)
		}

		// Notify other players
		h.broadcastToRoom(roomID, WebSocketMessage{
			Type:   "player_disconnected_timeout",
			RoomID: roomID,
			UserID: userID,
			Data: map[string]interface{}{
				"playerName": username,
				"message":    fmt.Sprintf("%s was disconnected and didn't reconnect in time", username),
			},
		})

		// Check if game should continue or end
		go h.checkGameCompletion(roomID)
	}
}

// handlePlayerReconnection handles when a player reconnects to an active game
func (h *Hub) handlePlayerReconnection(userID int, roomID string) {
	// Get player info
	username, err := h.getUsernameByID(userID)
	if err != nil {
		log.Printf("Failed to get username for reconnected player: %v", err)
		username = "Unknown Player"
	}

	// Update player status to active
	err = h.db.UpdatePlayerStatus(roomID, userID, "active")
	if err != nil {
		log.Printf("Failed to update player status to active: %v", err)
	}

	// Notify other players about reconnection
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "player_reconnected",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"playerName": username,
			"message":    fmt.Sprintf("%s has reconnected", username),
		},
	})

	log.Printf("Player %s reconnected to room %s", username, roomID)
}

// checkReconnection checks if a player is reconnecting to an active game
func (h *Hub) checkReconnection(userID int, roomID string) {
	// Get room info to check game status
	room, err := h.db.GetMultiplayerRoom(roomID)
	if err != nil {
		log.Printf("Failed to get room for reconnection check: %v", err)
		return
	}

	// Check if room is in active game state
	if room.Status == "active" {
		// Check if player was marked as disconnected
		for _, player := range room.Players {
			if player.UserID == userID && player.Status == "disconnected" {
				h.handlePlayerReconnection(userID, roomID)
				return
			}
		}
	}
}

// handlePlayerDisconnectMessage handles explicit disconnect messages from clients
func (h *Hub) handlePlayerDisconnectMessage(message WebSocketMessage) {
	userID := message.UserID
	roomID := message.RoomID

	// Extract reason from message data
	reason := "user_initiated"
	if message.Data != nil {
		if r, ok := message.Data["reason"].(string); ok {
			reason = r
		}
	}

	log.Printf("Received explicit disconnect message from user %d in room %s, reason: %s", userID, roomID, reason)

	// Call the existing disconnect handling logic
	h.handlePlayerDisconnection(userID, roomID)
}

// handleSpectateRequest handles requests to spectate an ongoing game
func (h *Hub) handleSpectateRequest(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		return
	}

	// Get room info to check if game is active
	room, err := h.db.GetMultiplayerRoom(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room for spectate request: %v", err)
		return
	}

	// Only allow spectating active games
	if room.Status != "active" {
		// Send error response
		h.sendToUser(message.UserID, WebSocketMessage{
			Type:  "spectate_error",
			Error: "Game is not currently active",
		})
		return
	}

	// Get current game state for all players
	gameStates, err := h.getActiveGameStates(message.RoomID)
	if err != nil {
		log.Printf("Failed to get game states for spectating: %v", err)
		return
	}

	// Get player usernames for display
	playerInfo := make(map[string]interface{})
	for _, player := range room.Players {
		if player.Status == "active" || player.Status == "finished" {
			playerInfo[fmt.Sprintf("player_%d", player.UserID)] = map[string]interface{}{
				"username": player.Username,
				"score":    player.Score,
				"status":   player.Status,
			}
		}
	}

	// Send spectate data
	h.sendToUser(message.UserID, WebSocketMessage{
		Type:   "spectate_data",
		RoomID: message.RoomID,
		Data: map[string]interface{}{
			"gameStates": gameStates,
			"playerInfo": playerInfo,
			"roomName":   room.Name,
			"gameType":   room.GameType,
		},
	})

	log.Printf("User %d started spectating room %s", message.UserID, message.RoomID)
}

// getActiveGameStates retrieves current game states for all active players
func (h *Hub) getActiveGameStates(roomID string) (map[string]interface{}, error) {
	// Get results for active game
	rows, err := h.db.GetGameResults(roomID)
	if err != nil {
		return nil, err
	}

	gameStates := make(map[string]interface{})
	// Note: For a complete implementation, you'd want to store and retrieve
	// actual game board states, but for now we'll return basic info
	for _, result := range rows {
		userID := result["userID"]
		gameStates[fmt.Sprintf("player_%v", userID)] = map[string]interface{}{
			"score":  result["score"],
			"status": "active", // Simplified for now
		}
	}

	return gameStates, nil
}

// sendToUser sends a message to a specific user
func (h *Hub) sendToUser(userID int, message WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
			break
		}
	}
}

// broadcastToRoom sends a message to all clients in a room
func (h *Hub) broadcastToRoom(roomID string, message WebSocketMessage) {
	h.mutex.RLock()
	roomClients := h.rooms[roomID]
	h.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	for client := range roomClients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			h.mutex.Lock()
			delete(h.clients, client)
			delete(h.rooms[roomID], client)
			h.mutex.Unlock()
		}
	}
}

// NotifyPlayerLeft notifies other players that someone left during an active game
func (h *Hub) NotifyPlayerLeft(roomID string, userID int, username string) {
	log.Printf("Notifying room %s that player %s left during active game", roomID, username)

	// Broadcast to all players in the room that the match has ended due to a player leaving
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "match_ended",
		RoomID: roomID,
		Data: map[string]interface{}{
			"reason":     "player_left",
			"playerName": username,
			"message":    fmt.Sprintf("Match ended: %s left the game", username),
		},
	})
}

// NotifyPlayerLeftWaiting notifies other players that someone left a waiting room
func (h *Hub) NotifyPlayerLeftWaiting(roomID string, userID int, username string) {
	log.Printf("Notifying room %s that player %s left waiting room", roomID, username)

	// Get updated room info to send to remaining players
	room, err := h.db.GetMultiplayerRoom(roomID)
	if err != nil {
		log.Printf("Failed to get room after player left: %v", err)
		return
	}

	// Broadcast room update to remaining players
	h.broadcastToRoom(roomID, WebSocketMessage{
		Type:   "room_update",
		RoomID: roomID,
		Data: map[string]interface{}{
			"room": room,
			"playerLeft": map[string]interface{}{
				"playerName": username,
				"message":    fmt.Sprintf("%s left the room", username),
			},
		},
	})
}

// ServeWS handles WebSocket requests
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract room ID from path
	roomID := r.PathValue("roomId")
	if roomID == "" {
		log.Printf("No room ID provided in WebSocket connection")
		conn.Close()
		return
	}

	// Get JWT token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("No token provided in WebSocket connection")
		conn.Close()
		return
	}

	// Validate JWT token and get user info
	userInfo, err := h.validateJWT(token)
	if err != nil {
		log.Printf("Invalid JWT token in WebSocket connection: %v", err)
		conn.Close()
		return
	}

	client := &Client{
		ID:     generateClientID(),
		UserID: userInfo.ID,
		RoomID: roomID,
		Conn:   conn,
		Send:   make(chan WebSocketMessage, 256),
		Hub:    h,
	}

	client.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Set user ID and room ID from client
		message.UserID = c.UserID
		message.RoomID = c.RoomID

		c.Hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
