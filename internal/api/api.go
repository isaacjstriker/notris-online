package api

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/isaacjstriker/devware/internal/config"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/web"
)

// APIServer represents the main server for the application
type APIServer struct {
	listenAddr string
	db         *database.DB
	config     *config.Config
}

// NewAPIServer creates a new APIServer instance
func NewAPIServer(listenAddr string, db *database.DB, config *config.Config) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		db:         db,
		config:     config,
	}
}

// Start runs the HTTP server
func (s *APIServer) Start() {
	router := http.NewServeMux()

	// --- Frontend Routes ---
	// Create a sub-filesystem for the static files
	staticFS, err := fs.Sub(web.Files, "static")
	if err != nil {
		log.Fatalf("could not create static filesystem: %v", err)
	}
	// Serve static files from the embedded filesystem
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Serve the main HTML file from the embedded filesystem
	router.HandleFunc("/", s.handleIndex)

	// --- API Routes ---
	router.HandleFunc("POST /api/register", s.handleRegister)
	router.HandleFunc("POST /api/login", s.handleLogin)
	router.HandleFunc("POST /api/logout", s.handleLogout)
	router.HandleFunc("GET /api/leaderboard/{gameType}", s.handleGetLeaderboard)
	router.HandleFunc("GET /api/recent/{gameType}", s.handleGetRecentGames)
	router.HandleFunc("POST /api/scores", s.handleSubmitScore)

	// --- WebSocket Route for Games ---
	router.HandleFunc("/ws/game", s.handleGameConnection)

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
	// Read the file from the embedded FS
	indexHTML, err := web.Files.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "could not read index file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}
