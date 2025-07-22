package tetris

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/eiannone/keyboard" // Add this import
	"github.com/isaacjstriker/devware/internal/auth"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/internal/types"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
	PreviewSize = 4
)

// Tetris represents a Tetris game instance
type Tetris struct {
	board        [][]int
	currentPiece *Piece
	nextPiece    *Piece
	score        int
	lines        int
	level        int
	gameOver     bool
	dropTimer    time.Time
	dropInterval time.Duration
}

// Piece represents a Tetris piece
type Piece struct {
	shape     [][]int
	x, y      int
	rotation  int
	pieceType int
}

// Tetris pieces (7 standard pieces)
var pieces = [][][]int{
	// I-piece
	{
		{1, 1, 1, 1},
	},
	// O-piece
	{
		{1, 1},
		{1, 1},
	},
	// T-piece
	{
		{0, 1, 0},
		{1, 1, 1},
	},
	// S-piece
	{
		{0, 1, 1},
		{1, 1, 0},
	},
	// Z-piece
	{
		{1, 1, 0},
		{0, 1, 1},
	},
	// J-piece
	{
		{1, 0, 0},
		{1, 1, 1},
	},
	// L-piece
	{
		{0, 0, 1},
		{1, 1, 1},
	},
}

var pieceColors = []string{"🟦", "🟨", "🟪", "🟩", "🟥", "🟧", "⬜"}

// NewTetris creates a new Tetris game
func NewTetris() *Tetris {
	t := &Tetris{
		board:        make([][]int, BoardHeight),
		score:        0,
		lines:        0,
		level:        1,
		dropInterval: time.Millisecond * 1000,
		dropTimer:    time.Now(),
	}

	// Initialize board
	for i := range t.board {
		t.board[i] = make([]int, BoardWidth)
	}

	t.spawnPiece()
	t.spawnNextPiece()

	return t
}

// spawnPiece creates a new random piece
func (t *Tetris) spawnPiece() {
	if t.nextPiece != nil {
		t.currentPiece = t.nextPiece
		t.currentPiece.x = BoardWidth/2 - 1
		t.currentPiece.y = 0
		t.spawnNextPiece() // Generate new next piece
	} else {
		pieceType := rand.Intn(len(pieces))
		t.currentPiece = &Piece{
			shape:     copyShape(pieces[pieceType]),
			x:         BoardWidth/2 - 1,
			y:         0,
			rotation:  0,
			pieceType: pieceType,
		}
	}

	// Check game over
	if t.checkCollision(t.currentPiece, 0, 0) {
		t.gameOver = true
	}
}

// spawnNextPiece creates the next piece for preview
func (t *Tetris) spawnNextPiece() {
	pieceType := rand.Intn(len(pieces))
	t.nextPiece = &Piece{
		shape:     copyShape(pieces[pieceType]),
		x:         0,
		y:         0,
		rotation:  0,
		pieceType: pieceType,
	}
}

// Play starts the Tetris game
func (t *Tetris) Play(db interface{}, authManager interface{}) *types.GameResult {
	var realDB *database.DB
	var realAuth *auth.CLIAuth

	if db != nil {
		var ok bool
		realDB, ok = db.(*database.DB)
		if !ok {
			return &types.GameResult{
				GameName: "Tetris",
				Score:    0,
				Duration: 0,
				Accuracy: 0,
				Perfect:  false,
				Bonus:    0,
				Metadata: map[string]interface{}{
					"error": "Invalid database connection",
				},
			}
		}
	}

	if authManager != nil {
		var ok bool
		realAuth, ok = authManager.(*auth.CLIAuth)
		if !ok {
			return &types.GameResult{
				GameName: "Tetris",
				Score:    0,
				Duration: 0,
				Accuracy: 0,
				Perfect:  false,
				Bonus:    0,
				Metadata: map[string]interface{}{
					"error": "Invalid authentication manager",
				},
			}
		}
	}

	fmt.Println("🧱 TETRIS - Stack blocks and clear lines!")
	fmt.Println("Controls: A/D = Move, S = Soft drop, W = Rotate, Q = Quit")
	fmt.Println("Press any key to start...")

	// Initialize keyboard for input
	if err := keyboard.Open(); err != nil {
		fmt.Printf("Failed to initialize keyboard: %v\n", err)
		return &types.GameResult{
			GameName: "Tetris",
			Score:    0,
			Duration: 0,
			Accuracy: 0,
			Perfect:  false,
			Bonus:    0,
			Metadata: map[string]interface{}{
				"error": "Failed to initialize keyboard input",
			},
		}
	}
	defer keyboard.Close()

	fmt.Scanln()
	startTime := time.Now()

	// Create a channel for input handling
	inputChan := make(chan bool, 1)
	quitChan := make(chan bool, 1)

	// Start input handler in a goroutine
	go t.inputHandler(inputChan, quitChan)

	gameLoop:
		for !t.gameOver {
			t.update()
			t.render()
	
			// Check for quit signal from input handler
			select {
			case quit := <-quitChan:
				if quit {
					break gameLoop
				}
			default:
				// Continue if no quit signal
			}
	
			// Game loop delay
			time.Sleep(50 * time.Millisecond)
		}

	// Signal input handler to stop
	close(inputChan)

	gameTime := time.Since(startTime)

	fmt.Println("\n🎮 GAME OVER!")
	fmt.Printf("📊 Final Score: %d\n", t.score)
	fmt.Printf("📏 Lines Cleared: %d\n", t.lines)

	// Calculate final score with bonuses
	finalScore := t.calculateFinalScore(gameTime)

	// Calculate accuracy based on lines cleared vs time played
	accuracy := float64(t.lines) / gameTime.Minutes() * 10 // rough accuracy metric
	if accuracy > 100 {
		accuracy = 100
	}

	// Check if it's a perfect game (high score with good efficiency)
	perfect := t.lines >= 40 && accuracy >= 80

	if realAuth != nil && realAuth.GetSession().IsLoggedIn() {
		session := realAuth.GetSession().GetCurrentSession()
		if session != nil && realDB != nil {
			// Create additional data for Tetris-specific stats
			additionalData := map[string]interface{}{
				"lines":     t.lines,
				"level":     t.level,
				"game_time": gameTime.Seconds(),
			}

			err := realDB.SaveGameScore(session.UserID, "tetris", finalScore, additionalData)
			if err != nil {
				fmt.Printf("Failed to save game score: %v\n", err)
			}
		}
	}

	return &types.GameResult{
		GameName: "Tetris",
		Score:    finalScore,
		Duration: gameTime.Seconds(),
		Accuracy: accuracy,
		Perfect:  perfect,
		Bonus:    finalScore - t.score, // difference between final and base score
		Metadata: map[string]interface{}{
			"lines":       t.lines,
			"level":       t.level,
			"base_score":  t.score,
			"time_bonus":  int(gameTime.Minutes()) * 50,
			"level_bonus": (t.level - 1) * 100,
		},
	}
}

func (t *Tetris) update() {
	// Drop piece automatically based on level
	if time.Since(t.dropTimer) > t.dropInterval {
		if !t.movePiece(0, 1) {
			t.placePiece()
			t.clearLines()
			t.spawnPiece() // This will handle next piece generation
		}
		t.dropTimer = time.Now()
	}
}

// render displays the game state
func (t *Tetris) render() {
	if t.currentPiece == nil {
		return
	}

	fmt.Print("\033[2J\033[H")

	fmt.Printf("🧱 TETRIS | Score: %d | Lines: %d | Level: %d\n", t.score, t.lines, t.level)
	fmt.Println(strings.Repeat("═", 50))

	// Create display board
	display := make([][]string, BoardHeight)
	for i := range display {
		display[i] = make([]string, BoardWidth)
		for j := range display[i] {
			if t.board[i][j] == 0 {
				display[i][j] = "⬛"
			} else {
				display[i][j] = pieceColors[t.board[i][j]-1]
			}
		}
	}

	// Add current piece to display
	if t.currentPiece != nil {
		for py := 0; py < len(t.currentPiece.shape); py++ {
			for px := 0; px < len(t.currentPiece.shape[py]); px++ {
				if t.currentPiece.shape[py][px] == 1 {
					boardY := t.currentPiece.y + py
					boardX := t.currentPiece.x + px
					if boardY >= 0 && boardY < BoardHeight && boardX >= 0 && boardX < BoardWidth {
						display[boardY][boardX] = pieceColors[t.currentPiece.pieceType]
					}
				}
			}
		}
	}

	// Print board with next piece preview
	for i, row := range display {
		fmt.Print("║")
		for _, cell := range row {
			fmt.Print(cell)
		}
		fmt.Print("║")

		// Show next piece preview on the right
		if i < 4 && t.nextPiece != nil && len(t.nextPiece.shape) > 0 {
			if i == 0 {
				fmt.Print("  Next:")
			}
			if i == 1 {
				fmt.Print("  ")
				for py := 0; py < len(t.nextPiece.shape) && py < 2; py++ {
					for px := 0; px < len(t.nextPiece.shape[py]) && px < 4; px++ {
						if t.nextPiece.shape[py][px] == 1 {
							fmt.Print(pieceColors[t.nextPiece.pieceType])
						} else {
							fmt.Print("⬛")
						}
					}
					if py == 0 {
						fmt.Print("\n" + strings.Repeat(" ", 12))
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("╚" + strings.Repeat("═", BoardWidth) + "╝")
	fmt.Println("Controls: A/D=Move, S=Down, W=Rotate, Q=Quit")
}

// inputHandler runs in a separate goroutine to handle input
func (t *Tetris) inputHandler(inputChan <-chan bool, quitChan chan<- bool) {
	for {
		select {
		case <-inputChan:
			return // Stop input handler
		default:
			char, key, err := keyboard.GetKey()
			if err != nil {
				time.Sleep(10 * time.Millisecond) // Small delay to prevent busy waiting
				continue
			}

			switch {
			case char == 'q' || char == 'Q':
				quitChan <- true
				return
			case char == 'a' || char == 'A':
				t.movePiece(-1, 0)
			case char == 'd' || char == 'D':
				t.movePiece(1, 0)
			case char == 's' || char == 'S':
				t.movePiece(0, 1)
			case char == 'w' || char == 'W':
				t.rotatePiece()
			case key == keyboard.KeyArrowLeft:
				t.movePiece(-1, 0)
			case key == keyboard.KeyArrowRight:
				t.movePiece(1, 0)
			case key == keyboard.KeyArrowDown:
				t.movePiece(0, 1)
			case key == keyboard.KeyArrowUp:
				t.rotatePiece()
			case key == keyboard.KeyEsc:
				quitChan <- true
				return
			}
		}
	}
}

// movePiece attempts to move the current piece
func (t *Tetris) movePiece(dx, dy int) bool {
	if t.currentPiece == nil {
		return false
	}

	if !t.checkCollision(t.currentPiece, dx, dy) {
		t.currentPiece.x += dx
		t.currentPiece.y += dy
		return true
	}
	return false
}

// rotatePiece rotates the current piece
func (t *Tetris) rotatePiece() {
	if t.currentPiece == nil {
		return
	}

	originalShape := t.currentPiece.shape
	t.currentPiece.shape = rotateShape(t.currentPiece.shape)

	if t.checkCollision(t.currentPiece, 0, 0) {
		t.currentPiece.shape = originalShape // Revert if collision
	}
}

// checkCollision checks if a piece would collide
func (t *Tetris) checkCollision(piece *Piece, dx, dy int) bool {
	for py := 0; py < len(piece.shape); py++ {
		for px := 0; px < len(piece.shape[py]); px++ {
			if piece.shape[py][px] == 1 {
				newX := piece.x + px + dx
				newY := piece.y + py + dy

				// Check boundaries
				if newX < 0 || newX >= BoardWidth || newY >= BoardHeight {
					return true
				}

				// Check board collision (ignore if above board)
				if newY >= 0 && t.board[newY][newX] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// placePiece places the current piece on the board
func (t *Tetris) placePiece() {
	if t.currentPiece == nil {
		return
	}

	for py := 0; py < len(t.currentPiece.shape); py++ {
		for px := 0; px < len(t.currentPiece.shape[py]); px++ {
			if t.currentPiece.shape[py][px] == 1 {
				boardY := t.currentPiece.y + py
				boardX := t.currentPiece.x + px
				if boardY >= 0 && boardY < BoardHeight && boardX >= 0 && boardX < BoardWidth {
					t.board[boardY][boardX] = t.currentPiece.pieceType + 1
				}
			}
		}
	}
}

// clearLines removes completed lines and updates score
func (t *Tetris) clearLines() {
	linesCleared := 0

	for y := BoardHeight - 1; y >= 0; y-- {
		fullLine := true
		for x := 0; x < BoardWidth; x++ {
			if t.board[y][x] == 0 {
				fullLine = false
				break
			}
		}

		if fullLine {
			// Remove line
			copy(t.board[1:y+1], t.board[0:y])
			t.board[0] = make([]int, BoardWidth)
			y++ // Check same line again
			linesCleared++
		}
	}

	if linesCleared > 0 {
		t.lines += linesCleared

		// Score calculation (similar to original Tetris)
		lineScores := []int{0, 40, 100, 300, 1200}
		if linesCleared < len(lineScores) {
			t.score += lineScores[linesCleared] * (t.level + 1)
		}

		// Level progression
		t.level = (t.lines / 10) + 1
		t.dropInterval = time.Millisecond * time.Duration(1000-(t.level-1)*50)
		if t.dropInterval < 50*time.Millisecond {
			t.dropInterval = 50 * time.Millisecond
		}
	}
}

// calculateFinalScore adds time and level bonuses
func (t *Tetris) calculateFinalScore(gameTime time.Duration) int {
	finalScore := t.score

	// Time bonus (bonus for lasting longer)
	timeBonus := int(gameTime.Minutes()) * 50
	finalScore += timeBonus

	// Level bonus
	levelBonus := (t.level - 1) * 100
	finalScore += levelBonus

	return finalScore
}

// Helper functions
func copyShape(original [][]int) [][]int {
	shape := make([][]int, len(original))
	for i := range original {
		shape[i] = make([]int, len(original[i]))
		copy(shape[i], original[i])
	}
	return shape
}

func rotateShape(shape [][]int) [][]int {
	rows := len(shape)
	cols := len(shape[0])
	rotated := make([][]int, cols)

	for i := range rotated {
		rotated[i] = make([]int, rows)
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			rotated[c][rows-1-r] = shape[r][c]
		}
	}

	return rotated
}

// GetName returns the game name
func (t *Tetris) GetName() string {
	return "Tetris"
}

// GetDescription returns the game description
func (t *Tetris) GetDescription() string {
	return "🧱 Classic block-stacking puzzle game. Clear lines by filling rows completely!"
}

// GetDifficulty returns the current difficulty level of the Tetris game
func (t *Tetris) GetDifficulty() int {
	return t.level
}

func (t *Tetris) IsAvailable() bool {
	return true
}
