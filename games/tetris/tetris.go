package tetris

import (
	"crypto/rand"
	"math/big"
	"os"
	"time"
)

// secureRandIntn generates a cryptographically secure random number in range [0, n)
func secureRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	result, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// Fallback to time-based randomness if crypto/rand fails
		return int(time.Now().UnixNano()) % n
	}
	return int(result.Int64())
}

const (
	BoardWidth  = 10
	BoardHeight = 20
)

type GameState struct {
	Board      [][]int `json:"board"`
	NextPiece  [][]int `json:"nextPiece"`
	HoldPiece  [][]int `json:"holdPiece"`
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
		TimePlayed   int     `json:"timePlayed"`
		PiecesPlaced int     `json:"piecesPlaced"`
		PPM          float64 `json:"ppm"`
		LineStats    [4]int  `json:"lineStats"`
	} `json:"stats"`
}

type Tetris struct {
	board         [][]int
	currentPiece  *Piece
	nextPiece     *Piece
	holdPiece     *Piece
	score         int
	lines         int
	level         int
	startingLevel int
	gameOver      bool
	paused        bool
	holdUsed      bool

	startTime     time.Time
	pausedTime    time.Duration
	lastPauseTime time.Time
	piecesPlaced  int
	lineStats     [4]int // [singles, doubles, triples, tetris]

	// For web socket communication
	dropCounter int
	dropSpeed   int
}

type Piece struct {
	shape     [][]int
	x, y      int
	rotation  int
	pieceType int
}

var pieces = [][][]int{
	{
		{1, 1, 1, 1},
	},
	{
		{1, 1},
		{1, 1},
	},
	{
		{0, 1, 0},
		{1, 1, 1},
	},
	{
		{0, 1, 1},
		{1, 1, 0},
	},
	{
		{1, 1, 0},
		{0, 1, 1},
	},
	{
		{1, 0, 0},
		{1, 1, 1},
	},
	{
		{0, 0, 1},
		{1, 1, 1},
	},
}

var pieceColors = []string{"##", "@@", "**", "%%", "&&", "++", "=="}

func NewTetris() *Tetris {
	t := &Tetris{
		board:         make([][]int, BoardHeight),
		score:         0,
		lines:         0,
		level:         1,
		gameOver:      false,
		startTime:     time.Now(),
		pausedTime:    0,
		lastPauseTime: time.Now(),
		piecesPlaced:  0,
		holdUsed:      false,
	}

	for i := range t.board {
		t.board[i] = make([]int, BoardWidth)
	}

	t.dropSpeed = t.getFramesPerDrop()

	t.spawnPiece()
	t.spawnNextPiece()

	return t
}

func (t *Tetris) spawnPiece() {
	if t.nextPiece != nil {
		t.currentPiece = t.nextPiece
		t.currentPiece.x = BoardWidth/2 - 1
		t.currentPiece.y = 0
		t.spawnNextPiece()
	} else {
		pieceType := secureRandIntn(len(pieces))
		t.currentPiece = &Piece{
			shape:     copyShape(pieces[pieceType]),
			x:         BoardWidth/2 - 1,
			y:         0,
			rotation:  0,
			pieceType: pieceType,
		}
	}

	t.holdUsed = false

	if t.checkCollision(t.currentPiece, 0, 0) {
		t.gameOver = true
	}
}

func (t *Tetris) spawnNextPiece() {
	pieceType := secureRandIntn(len(pieces))
	t.nextPiece = &Piece{
		shape:     copyShape(pieces[pieceType]),
		x:         0,
		y:         0,
		rotation:  0,
		pieceType: pieceType,
	}
}

func (t *Tetris) GetState() GameState {
	boardCopy := make([][]int, BoardHeight)
	for i := range t.board {
		boardCopy[i] = make([]int, BoardWidth)
		copy(boardCopy[i], t.board[i])
	}

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

	var holdPieceShape [][]int
	if t.holdPiece != nil {
		holdPieceShape = t.holdPiece.shape
	}

	ghostShape, ghostX, ghostY := t.calculateGhostPiece()

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
		HoldPiece: holdPieceShape,
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
	case "hold":
		if !t.paused {
			t.holdCurrentPiece()
		}
	case "pause":
		t.togglePause()
	}
}

func (t *Tetris) IsGameOver() bool {
	return t.gameOver
}

func (t *Tetris) GetScore() int {
	return t.score
}

func (t *Tetris) Update() {
	if t.gameOver || t.paused {
		return
	}

	t.dropCounter++
	if t.dropCounter >= t.dropSpeed {
		t.dropCounter = 0
		if !t.movePiece(0, 1) {
			t.placePiece()
			t.clearLines()
			t.spawnPiece()
		}
	}
}

func (t *Tetris) togglePause() {
	if !t.gameOver {
		if t.paused {
			if !t.lastPauseTime.IsZero() {
				t.pausedTime += time.Since(t.lastPauseTime)
				t.lastPauseTime = time.Time{}
			}
		} else {
			t.lastPauseTime = time.Now()
		}
		t.paused = !t.paused
	}
}

func init() {
	if supportsColor() {
		pieceColors = []string{
			"\033[44m  \033[0m",
			"\033[43m  \033[0m",
			"\033[45m  \033[0m",
			"\033[42m  \033[0m",
			"\033[41m  \033[0m",
			"\033[47m  \033[0m",
			"\033[46m  \033[0m",
		}
	} else {
		pieceColors = []string{"##", "@@", "**", "%%", "&&", "++", "=="}
	}
}

func supportsColor() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

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

func (t *Tetris) rotatePiece() {
	if t.currentPiece == nil {
		return
	}

	originalShape := t.currentPiece.shape
	rotatedShape := rotateShape(t.currentPiece.shape)
	t.currentPiece.shape = rotatedShape

	if !t.checkCollision(t.currentPiece, 0, 0) {
		return
	}

	kickOffsets := []int{1, -1, 2, -2}
	for _, dx := range kickOffsets {
		if !t.checkCollision(t.currentPiece, dx, 0) {
			t.currentPiece.x += dx
			return
		}
	}

	t.currentPiece.shape = originalShape
}

func (t *Tetris) hardDrop() {
	if t.currentPiece == nil {
		return
	}

	for t.movePiece(0, 1) {
	}

	t.placePiece()
}

func (t *Tetris) holdCurrentPiece() {
	if t.currentPiece == nil || t.holdUsed {
		return
	}

	if t.holdPiece == nil {
		t.holdPiece = &Piece{
			shape:     copyShape(pieces[t.currentPiece.pieceType]),
			x:         0,
			y:         0,
			rotation:  0,
			pieceType: t.currentPiece.pieceType,
		}
		t.spawnPiece()
	} else {
		heldPieceType := t.holdPiece.pieceType

		t.holdPiece = &Piece{
			shape:     copyShape(pieces[t.currentPiece.pieceType]),
			x:         0,
			y:         0,
			rotation:  0,
			pieceType: t.currentPiece.pieceType,
		}

		t.currentPiece = &Piece{
			shape:     copyShape(pieces[heldPieceType]),
			x:         BoardWidth/2 - 1,
			y:         0,
			rotation:  0,
			pieceType: heldPieceType,
		}

		if t.checkCollision(t.currentPiece, 0, 0) {
			t.gameOver = true
			return
		}
	}

	t.holdUsed = true
}

func (t *Tetris) calculateGhostPiece() ([][]int, int, int) {
	if t.currentPiece == nil {
		return nil, 0, 0
	}

	ghostX := t.currentPiece.x
	ghostY := t.currentPiece.y

	for {
		if t.checkCollisionAt(t.currentPiece.shape, ghostX, ghostY+1) {
			break
		}
		ghostY++
	}

	return t.currentPiece.shape, ghostX, ghostY
}

func (t *Tetris) checkCollisionAt(shape [][]int, x, y int) bool {
	for py := 0; py < len(shape); py++ {
		for px := 0; px < len(shape[py]); px++ {
			if shape[py][px] == 1 {
				newX := x + px
				newY := y + py

				if newX < 0 || newX >= BoardWidth || newY >= BoardHeight {
					return true
				}

				if newY >= 0 && t.board[newY][newX] != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (t *Tetris) checkCollision(piece *Piece, dx, dy int) bool {
	for py := 0; py < len(piece.shape); py++ {
		for px := 0; px < len(piece.shape[py]); px++ {
			if piece.shape[py][px] == 1 {
				newX := piece.x + px + dx
				newY := piece.y + py + dy

				if newX < 0 || newX >= BoardWidth || newY >= BoardHeight {
					return true
				}

				if newY >= 0 && t.board[newY][newX] != 0 {
					return true
				}
			}
		}
	}
	return false
}

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

	t.piecesPlaced++
}

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
			copy(t.board[1:y+1], t.board[0:y])
			t.board[0] = make([]int, BoardWidth)
			y++
			linesCleared++
		}
	}

	if linesCleared > 0 {
		t.lines += linesCleared

		if linesCleared >= 1 && linesCleared <= 4 {
			t.lineStats[linesCleared-1]++
		}

		lineScores := []int{0, 40, 100, 300, 1200}
		if linesCleared < len(lineScores) {
			t.score += lineScores[linesCleared] * (t.level + 1)
		}

		newLevel := t.startingLevel + (t.lines / 10)
		if newLevel != t.level {
			t.level = newLevel
			t.dropSpeed = t.getFramesPerDrop()
		}
	}
}

func (t *Tetris) getFramesPerDrop() int {
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

func (t *Tetris) SetLevel(level int) {
	if level >= 1 && level <= 29 {
		t.level = level
		t.startingLevel = level
		t.dropSpeed = t.getFramesPerDrop()
	}
}

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
