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
		return true
	},
}

func (s *APIServer) handleGameConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	roomID := r.URL.Query().Get("room")
	isMultiplayer := r.URL.Query().Get("multiplayer") == "true"

	game := tetris.NewTetris()

	if isMultiplayer && roomID != "" {
		room, err := s.db.GetMultiplayerRoom(roomID)
		if err == nil && room.Settings != nil {
			if startingLevel, ok := room.Settings["starting_level"]; ok {
				if level, ok := startingLevel.(float64); ok {
					game.SetLevel(int(level))
					log.Printf("Set multiplayer game starting level to %d for room %s", int(level), roomID)
				}
			}
		}
	}

	gameLoop(conn, game)
}

func gameLoop(conn *websocket.Conn, game *tetris.Tetris) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	inputChan := make(chan string)
	go func() {
		defer close(inputChan)
		for {
			var msg struct {
				Type  string `json:"type"`
				Key   string `json:"key"`
				Level int    `json:"level"`
			}
			err := conn.ReadJSON(&msg)
			if err != nil {
				return
			}
			switch msg.Type {
case "input":
				inputChan <- msg.Key
			case "setLevel":
				game.SetLevel(msg.Level)
			}
		}
	}()

	for {
		select {
		case input, ok := <-inputChan:
			if !ok {
				return
			}
			game.HandleWebInput(input)

		case <-ticker.C:
			if game.IsGameOver() {
				conn.WriteJSON(map[string]interface{}{"type": "gameOver", "score": game.GetScore()})
				return
			}
			game.Update()

			err := conn.WriteJSON(game.GetState())
			if err != nil {
				return
			}
		}
	}
}
