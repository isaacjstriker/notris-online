package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/isaacjstriker/devware/games/tetris"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for now. In production, you'd want to restrict this.
		return true
	},
}

// handleGameConnection upgrades an HTTP request to a WebSocket connection and starts a game.
func (s *APIServer) handleGameConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// For now, we'll hardcode the Tetris game.
	// Later, you can pass the game type in the URL.
	game := tetris.NewTetris()
	gameLoop(conn, game)
}

// gameLoop runs the main loop for a single game instance.
func gameLoop(conn *websocket.Conn, game *tetris.Tetris) {
	// Ticker to send game state updates to the client
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	// Goroutine to read messages from the client (e.g., keyboard input)
	inputChan := make(chan string)
	go func() {
		defer close(inputChan)
		for {
			var msg struct {
				Type string `json:"type"`
				Key  string `json:"key"`
			}
			err := conn.ReadJSON(&msg)
			if err != nil {
				// Client disconnected
				return
			}
			if msg.Type == "input" {
				inputChan <- msg.Key
			}
		}
	}()

	for {
		select {
		case input, ok := <-inputChan:
			if !ok {
				// Client disconnected
				return
			}
			// Process player input
			game.HandleWebInput(input)

		case <-ticker.C:
			// Update game state
			if game.IsGameOver() {
				// Notify client and close connection
				conn.WriteJSON(map[string]interface{}{"type": "gameOver", "score": game.GetScore()})
				return
			}
			game.Update()

			// Send the new game state to the client
			err := conn.WriteJSON(game.GetState())
			if err != nil {
				// Client disconnected
				return
			}
		}
	}
}
