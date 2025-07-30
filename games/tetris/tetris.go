package tetris

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
)

// GameState represents the data sent to the client for rendering.
type GameState struct {
	Board      [][]int `json:"board"`
	NextPiece  [][]int `json:"nextPiece"`
	GhostPiece struct {
		Shape [][]int `json:"shape"`
		X     int     `json:"x"`
		Y     int     `json:"y"`
	} `json:"ghostPiece"`
	Score    int  `json:"score"`
	Lines    int  `json:"lines"`
	Level    int  `json:"level"`
	GameOver bool `json:"gameOver"`
	Paused   bool `json:"paused"`
	Stats    struct {
		TimePlayed   int     `json:"timePlayed"` // seconds
		PiecesPlaced int     `json:"piecesPlaced"`
		PPM          float64 `json:"ppm"`       // pieces per minute
		LineStats    [4]int  `json:"lineStats"` // [singles, doubles, triples, tetris]
	} `json:"stats"`
}

// Tetris represents a Tetris game instance
type Tetris struct {
	board         [][]int
	currentPiece  *Piece
	nextPiece     *Piece
	score         int
	lines         int
	level         int
	startingLevel int
	gameOver      bool
	paused        bool

	// Game statistics
	startTime     time.Time
	pausedTime    time.Duration // Total time spent paused
	lastPauseTime time.Time     // When the current pause started
	piecesPlaced  int
	lineStats     [4]int // [singles, doubles, triples, tetris]

	// For web socket communication
	dropCounter int
	dropSpeed   int
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

var pieceColors = []string{"##", "@@", "**", "%%", "&&", "++", "=="}

// NewTetris creates a new Tetris game
func NewTetris() *Tetris {
	t := &Tetris{
		board:         make([][]int, BoardHeight),
		score:         0,
		lines:         0,
		level:         1,
		startingLevel: 1,
		dropCounter:   0,
		startTime:     time.Now(),
		piecesPlaced:  0,
	}

	// Initialize board
	for i := range t.board {
		t.board[i] = make([]int, BoardWidth)
	}

	// Set initial drop speed based on level 1
	t.dropSpeed = t.getFramesPerDrop()

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

// GetState returns the current state of the game for JSON serialization.
func (t *Tetris) GetState() GameState {
	// Create a copy of the board to draw the current piece on
	boardCopy := make([][]int, BoardHeight)
	for i := range t.board {
		boardCopy[i] = make([]int, BoardWidth)
		copy(boardCopy[i], t.board[i])
	}

	// Draw the current piece onto the board copy
	if t.currentPiece != nil {
		for py := 0; py < len(t.currentPiece.shape); py++ {
			for px := 0; px < len(t.currentPiece.shape[py]); px++ {
				if t.currentPiece.shape[py][px] == 1 {
					boardY := t.currentPiece.y + py
					boardX := t.currentPiece.x + px
					if boardY >= 0 && boardY < BoardHeight && boardX >= 0 && boardX < BoardWidth {
						boardCopy[boardY][boardX] = t.currentPiece.pieceType + 1
					}
				}
			}
		}
	}

	var nextPieceShape [][]int
	if t.nextPiece != nil {
		nextPieceShape = t.nextPiece.shape
	}

	// Calculate ghost piece position
	ghostShape, ghostX, ghostY := t.calculateGhostPiece()

	// Calculate statistics
	totalElapsed := time.Since(t.startTime)
	currentPauseTime := time.Duration(0)
	if t.paused && !t.lastPauseTime.IsZero() {
		currentPauseTime = time.Since(t.lastPauseTime)
	}
	actualPlayTime := totalElapsed - t.pausedTime - currentPauseTime
	timePlayed := int(actualPlayTime.Seconds())

	var ppm float64
	if timePlayed > 0 {
		ppm = float64(t.piecesPlaced) / (float64(timePlayed) / 60.0)
	}

	return GameState{
		Board:     boardCopy,
		NextPiece: nextPieceShape,
		GhostPiece: struct {
			Shape [][]int `json:"shape"`
			X     int     `json:"x"`
			Y     int     `json:"y"`
		}{
			Shape: ghostShape,
			X:     ghostX,
			Y:     ghostY,
		},
		Score:    t.score,
		Lines:    t.lines,
		Level:    t.level,
		GameOver: t.gameOver,
		Paused:   t.paused,
		Stats: struct {
			TimePlayed   int     `json:"timePlayed"`
			PiecesPlaced int     `json:"piecesPlaced"`
			PPM          float64 `json:"ppm"`
			LineStats    [4]int  `json:"lineStats"`
		}{
			TimePlayed:   timePlayed,
			PiecesPlaced: t.piecesPlaced,
			PPM:          ppm,
			LineStats:    t.lineStats,
		},
	}
}

// HandleWebInput processes a single command from the web client.
func (t *Tetris) HandleWebInput(input string) {
	switch input {
	case "left":
		if !t.paused {
			t.movePiece(-1, 0)
		}
	case "right":
		if !t.paused {
			t.movePiece(1, 0)
		}
	case "down":
		if !t.paused {
			t.movePiece(0, 1)
		}
	case "rotate":
		if !t.paused {
			t.rotatePiece()
		}
	case "hardDrop":
		if !t.paused {
			t.hardDrop()
		}
	case "pause":
		t.togglePause()
	}
}

// IsGameOver checks if the game has ended.
func (t *Tetris) IsGameOver() bool {
	return t.gameOver
}

// GetScore returns the final score.
func (t *Tetris) GetScore() int {
	return t.score
}

// Update is the main game tick function, replacing the old game loop.
func (t *Tetris) Update() {
	if t.gameOver || t.paused {
		return
	}

	t.dropCounter++
	if t.dropCounter >= t.dropSpeed {
		t.dropCounter = 0
		// Try to move the piece down
		if !t.movePiece(0, 1) {
			// If it fails, the piece has landed
			t.placePiece()
			t.clearLines()
			t.spawnPiece()
			// The spawnPiece function will set gameOver if a new piece can't be placed.
		}
	}
}

// togglePause toggles the pause state
func (t *Tetris) togglePause() {
	if !t.gameOver {
		if t.paused {
			// Resuming: add the pause duration to total paused time
			if !t.lastPauseTime.IsZero() {
				t.pausedTime += time.Since(t.lastPauseTime)
				t.lastPauseTime = time.Time{} // Reset
			}
		} else {
			// Pausing: record when the pause started
			t.lastPauseTime = time.Now()
		}
		t.paused = !t.paused
	}
}

// render displays the game state
func (t *Tetris) render() {
	if t.currentPiece == nil {
		return
	}

	fmt.Print("\033[2J\033[H")

	fmt.Printf("TETRIS | Score: %d | Lines: %d | Level: %d\n", t.score, t.lines, t.level)
	fmt.Println(strings.Repeat("═", 50))

	// Create display board
	display := make([][]string, BoardHeight)
	for i := range display {
		display[i] = make([]string, BoardWidth)
		for j := range display[i] {
			if t.board[i][j] == 0 {
				if supportsColor() {
					display[i][j] = "  " // Two spaces for color mode
				} else {
					display[i][j] = ".." // Dots for ASCII mode
				}
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

	// Print the main board
	for _, row := range display {
		fmt.Print("║")
		for _, cell := range row {
			fmt.Print(cell)
		}
		fmt.Print("║")
		fmt.Println()
	}

	// Fix the bottom border width
	fmt.Println("╚" + strings.Repeat("═", BoardWidth*2) + "╝")

	// Show next piece separately below the board
	if t.nextPiece != nil {
		fmt.Println("\nNext Piece:")
		for py := 0; py < len(t.nextPiece.shape); py++ {
			fmt.Print("  ")
			for px := 0; px < len(t.nextPiece.shape[py]); px++ {
				if t.nextPiece.shape[py][px] == 1 {
					fmt.Print(pieceColors[t.nextPiece.pieceType])
				} else {
					fmt.Print("  ")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\nControls: A/D=Move, S=Down, W=Rotate, Q=Quit")
}

func init() {
	// Check if terminal supports colors
	if supportsColor() {
		pieceColors = []string{
			"\033[44m  \033[0m", // Blue
			"\033[43m  \033[0m", // Yellow
			"\033[45m  \033[0m", // Magenta
			"\033[42m  \033[0m", // Green
			"\033[41m  \033[0m", // Red
			"\033[47m  \033[0m", // White
			"\033[46m  \033[0m", // Cyan
		}
	} else {
		pieceColors = []string{"##", "@@", "**", "%%", "&&", "++", "=="}
	}
}

func supportsColor() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
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

// rotatePiece rotates the current piece, with wall-kicking.
func (t *Tetris) rotatePiece() {
	if t.currentPiece == nil {
		return
	}

	originalShape := t.currentPiece.shape
	rotatedShape := rotateShape(t.currentPiece.shape)
	t.currentPiece.shape = rotatedShape

	// Try standard rotation
	if !t.checkCollision(t.currentPiece, 0, 0) {
		return // Success
	}

	// Wall kick tests (try moving left/right)
	// These are common offsets for I, J, L, S, T, Z pieces
	kickOffsets := []int{1, -1, 2, -2}
	for _, dx := range kickOffsets {
		if !t.checkCollision(t.currentPiece, dx, 0) {
			t.currentPiece.x += dx
			return // Success
		}
	}

	// If all kicks fail, revert
	t.currentPiece.shape = originalShape
}

// hardDrop drops the current piece to the bottom instantly
func (t *Tetris) hardDrop() {
	if t.currentPiece == nil {
		return
	}

	// Keep moving down until collision
	for t.movePiece(0, 1) {
		// Continue dropping
	}

	// Immediately place the piece
	t.placePiece()
}

// calculateGhostPiece calculates where the current piece would land
func (t *Tetris) calculateGhostPiece() ([][]int, int, int) {
	if t.currentPiece == nil {
		return nil, 0, 0
	}

	// Create a copy of the current piece
	ghostX := t.currentPiece.x
	ghostY := t.currentPiece.y

	// Drop the ghost piece down until it would collide
	for {
		// Check if moving down one more would cause collision
		if t.checkCollisionAt(t.currentPiece.shape, ghostX, ghostY+1) {
			break
		}
		ghostY++
	}

	return t.currentPiece.shape, ghostX, ghostY
}

// checkCollisionAt checks if a shape would collide at specific coordinates
func (t *Tetris) checkCollisionAt(shape [][]int, x, y int) bool {
	for py := 0; py < len(shape); py++ {
		for px := 0; px < len(shape[py]); px++ {
			if shape[py][px] == 1 {
				newX := x + px
				newY := y + py

				// Check boundaries
				if newX < 0 || newX >= BoardWidth || newY >= BoardHeight {
					return true
				}

				// Check collision with existing pieces (but not if above board)
				if newY >= 0 && t.board[newY][newX] != 0 {
					return true
				}
			}
		}
	}
	return false
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

	// Track statistics
	t.piecesPlaced++
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

		// Track line statistics
		if linesCleared >= 1 && linesCleared <= 4 {
			t.lineStats[linesCleared-1]++
		}

		// Score calculation (similar to original Tetris)
		lineScores := []int{0, 40, 100, 300, 1200}
		if linesCleared < len(lineScores) {
			t.score += lineScores[linesCleared] * (t.level + 1)
		}

		// Level progression - increase level every 10 lines from starting level
		newLevel := t.startingLevel + (t.lines / 10)
		if newLevel != t.level {
			t.level = newLevel
			// Update drop speed using traditional Tetris frames-per-drop system
			t.dropSpeed = t.getFramesPerDrop()
		}
	}
}

// getFramesPerDrop returns the number of game ticks before a piece drops one row
// This implements the traditional Tetris speed curve
func (t *Tetris) getFramesPerDrop() int {
	// Traditional Tetris speed table (frames per drop)
	speedTable := []int{
		48, // Level 1
		43, // Level 2
		38, // Level 3
		33, // Level 4
		28, // Level 5
		23, // Level 6
		18, // Level 7
		13, // Level 8
		8,  // Level 9
		6,  // Level 10
		5,  // Level 11-12
		5,  // Level 13-15
		4,  // Level 16-18
		4,  // Level 19-28
		3,  // Level 29+
	}

	if t.level <= 0 {
		return speedTable[0]
	} else if t.level <= 10 {
		return speedTable[t.level-1]
	} else if t.level <= 12 {
		return 5
	} else if t.level <= 15 {
		return 5
	} else if t.level <= 18 {
		return 4
	} else if t.level <= 28 {
		return 4
	} else {
		return 3 // Level 29+, maximum speed
	}
}

// SetLevel allows setting the starting level
// SetLevel allows setting the starting level
func (t *Tetris) SetLevel(level int) {
	if level >= 1 && level <= 29 {
		t.level = level
		t.startingLevel = level
		t.dropSpeed = t.getFramesPerDrop()
	}
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
