package multiplayer

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan WebSocketMessage
	register   chan *Client
	unregister chan *Client
	db         *database.DB
	mutex      sync.RWMutex
	stopCleanup chan bool
}

// NewHub creates a new WebSocket hub
func NewHub(db *database.DB) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		rooms:       make(map[string]map[*Client]bool),
		broadcast:   make(chan WebSocketMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		db:          db,
		stopCleanup: make(chan bool),
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
				if client.RoomID != "" && h.rooms[client.RoomID] != nil {
					delete(h.rooms[client.RoomID], client)
					if len(h.rooms[client.RoomID]) == 0 {
						delete(h.rooms, client.RoomID)
					}
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

			if roomClients != nil {
				for client := range roomClients {
					select {
					case client.Send <- WebSocketMessage{
						Type:  "room_closed",
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
	case "player_finished":
		h.handlePlayerFinished(message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
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

// handlePlayerFinished marks player as finished
func (h *Hub) handlePlayerFinished(message WebSocketMessage) {
	if message.RoomID == "" || message.UserID == 0 {
		return
	}

	score := 0
	if s, ok := message.Data["score"].(float64); ok {
		score = int(s)
	}

	position := 0
	if p, ok := message.Data["position"].(float64); ok {
		position = int(p)
	}

	err := h.db.FinishPlayerGame(message.RoomID, message.UserID, score, position)
	if err != nil {
		log.Printf("Failed to finish player game: %v", err)
		return
	}

	// Broadcast player finished
	h.broadcastToRoom(message.RoomID, WebSocketMessage{
		Type:   "player_finished",
		RoomID: message.RoomID,
		UserID: message.UserID,
		Data:   message.Data,
	})
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

// ServeWS handles WebSocket requests
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get user ID and room ID from query params
	userIDStr := r.URL.Query().Get("user_id")
	roomID := r.URL.Query().Get("room_id")

	// For now, we'll extract user ID from the token in a real implementation
	// This is a simplified version
	var userID int
	if userIDStr != "" {
		// In production, validate the user ID from JWT token
		// userID = validateAndGetUserID(r)
	}

	client := &Client{
		ID:     generateClientID(),
		UserID: userID,
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
