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
	Board     [][]int `json:"board"`
	NextPiece [][]int `json:"nextPiece"`
	Score     int     `json:"score"`
	Lines     int     `json:"lines"`
	Level     int     `json:"level"`
	GameOver  bool    `json:"gameOver"`
}

// Tetris represents a Tetris game instance
type Tetris struct {
	board        [][]int
	currentPiece *Piece
	nextPiece    *Piece
	score        int
	lines        int
	level        int
	gameOver     bool

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
		board:       make([][]int, BoardHeight),
		score:       0,
		lines:       0,
		level:       1,
		dropSpeed:   20, // Ticks per drop
		dropCounter: 0,
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

	return GameState{
		Board:     boardCopy,
		NextPiece: nextPieceShape,
		Score:     t.score,
		Lines:     t.lines,
		Level:     t.level,
		GameOver:  t.gameOver,
	}
}

// HandleWebInput processes a single command from the web client.
func (t *Tetris) HandleWebInput(input string) {
	switch input {
	case "left":
		t.movePiece(-1, 0)
	case "right":
		t.movePiece(1, 0)
	case "down":
		t.movePiece(0, 1)
	case "rotate":
		t.rotatePiece()
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
	if t.gameOver {
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
		// Adjust drop speed based on level
		newDropSpeed := 20 - (t.level - 1)
		if newDropSpeed < 2 {
			newDropSpeed = 2 // Set a minimum drop speed
		}
		t.dropSpeed = newDropSpeed
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
