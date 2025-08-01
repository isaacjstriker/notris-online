package api

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/isaacjstriker/devware/internal/config"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/internal/multiplayer"
	"github.com/isaacjstriker/devware/web"
)

type APIServer struct {
	listenAddr string
	db         *database.DB
	config     *config.Config
	wsHub      *multiplayer.Hub
}

func NewAPIServer(cfg *config.Config, db *database.DB) *APIServer {
	server := &APIServer{
		config:     cfg,
		db:         db,
		listenAddr: fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
	}

	// Create JWT validator function
	jwtValidator := func(tokenString string) (*multiplayer.UserInfo, error) {
		userInfo, err := server.validateJWT(tokenString)
		if err != nil {
			return nil, err
		}
		return &multiplayer.UserInfo{
			ID:       userInfo.UserID,
			Username: userInfo.Username,
		}, nil
	}

	server.wsHub = multiplayer.NewHub(db, jwtValidator)
	return server
}

func (s *APIServer) Start() {
	router := http.NewServeMux()

	staticFS, err := fs.Sub(web.Files, "static")
	if err != nil {
		log.Fatalf("could not create static filesystem: %v", err)
	}
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	router.HandleFunc("/", s.handleIndex)

	router.HandleFunc("POST /api/register", s.handleRegister)
	router.HandleFunc("POST /api/login", s.handleLogin)
	router.HandleFunc("POST /api/logout", s.handleLogout)
	router.HandleFunc("GET /api/leaderboard/{gameType}", s.handleGetLeaderboard)
	router.HandleFunc("GET /api/recent/{gameType}", s.handleGetRecentGames)
	router.HandleFunc("POST /api/scores", s.handleSubmitScore)

	router.HandleFunc("POST /api/rooms", requireAuth(s, s.handleCreateRoom))
	router.HandleFunc("GET /api/rooms/{gameType}", s.handleGetAvailableRooms)
	router.HandleFunc("GET /api/room/{roomId}", s.handleGetRoom)
	router.HandleFunc("POST /api/room/{roomId}/join", requireAuth(s, s.handleJoinRoom))
	router.HandleFunc("POST /api/room/{roomId}/leave", requireAuth(s, s.handleLeaveRoom))
	router.HandleFunc("POST /api/room/{roomId}/ready", requireAuth(s, s.handlePlayerReady))

	router.HandleFunc("GET /ws/room/{roomId}", s.handleWebSocket)
	router.HandleFunc("GET /ws/game", s.handleGameConnection)

	go s.wsHub.Run()

	log.Printf("API server listening on %s", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		log.Fatalf("could not start server: %s", err)
	}
}

// handleIndex serves the main index.html file from the embedded filesystem
func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	indexHTML, err := web.Files.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "could not read index file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}
