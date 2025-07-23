package breakout

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/isaacjstriker/devware/internal/auth"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/internal/types"
	lua "github.com/yuin/gopher-lua"
)

const (
	BoardWidth  = 40
	BoardHeight = 20
)

// Paddle represents the player's paddle
type Paddle struct {
	X, Y int
	Size int
}

// Ball represents the game ball
type Ball struct {
	X, Y   float64
	VX, VY float64 // Velocity
}

// Brick represents a single brick
type Brick struct {
	X, Y  int
	Alive bool
}

// GameConfig holds settings loaded from the Lua script
type GameConfig struct {
	PaddleSize  int
	BallSpeedX  float64
	BallSpeedY  float64
	Lives       int
	BrickScore  int
	BrickLayout [][]int // Add this field
}

// Game holds the state of the Breakout game
type Game struct {
	paddle    Paddle
	ball      Ball
	bricks    []Brick
	score     int
	lives     int
	level     int // Add level tracking
	gameOver  bool
	win       bool
	inputChan chan keyboard.Key
	config    *GameConfig
}

// NewBreakout creates a new Breakout game instance
func NewBreakout() *Game {
	config := loadConfig() // Load config from Lua

	g := &Game{
		paddle: Paddle{
			X:    BoardWidth / 2,
			Y:    BoardHeight - 2,
			Size: config.PaddleSize,
		},
		ball: Ball{
			X:  float64(BoardWidth / 2),
			Y:  float64(BoardHeight / 2),
			VX: config.BallSpeedX,
			VY: config.BallSpeedY,
		},
		score:     0,
		lives:     config.Lives,
		level:     1, // Start at level 1
		inputChan: make(chan keyboard.Key),
		config:    config,
	}

	g.resetBricks() // Create the initial set of bricks

	return g
}

// resetBricks clears and repopulates the brick layout for a new level.
func (g *Game) resetBricks() {
	g.bricks = []Brick{} // Clear existing bricks

	const brickWidth = 3
	const brickSpacing = 1
	const horizontalOffset = 2 // To center the layout

	for y, row := range g.config.BrickLayout {
		for x, cell := range row {
			if cell == 1 {
				brickX := horizontalOffset + x*(brickWidth+brickSpacing)
				brickY := y + 2 // Vertical offset
				g.bricks = append(g.bricks, Brick{X: brickX, Y: brickY, Alive: true})
			}
		}
	}
}

// Interface methods
func (g *Game) GetName() string {
	return "Breakout"
}

func (g *Game) GetDescription() string {
	return "Classic brick-breaking game. Don't let the ball fall!"
}

func (g *Game) GetDifficulty() int {
	return 2 // Medium
}

func (g *Game) IsAvailable() bool {
	return true
}

// Play starts the Breakout game
func (g *Game) Play(db interface{}, authManager interface{}) *types.GameResult {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("Failed to initialize keyboard: %v\n", err)
		return &types.GameResult{GameName: "Breakout", Score: -1}
	}
	defer keyboard.Close()

	// Start input handler in a goroutine
	go g.inputHandler()

	fmt.Println("--- BREAKOUT ---")
	fmt.Println("Controls: A/D or Left/Right Arrows to move. Q to quit.")
	time.Sleep(2 * time.Second)

	startTime := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond) // Use a ticker for a consistent frame rate
	defer ticker.Stop()

	for range ticker.C {
		g.processInput()
		g.update()
		g.render()
		if g.gameOver {
			break
		}
	}

	duration := time.Since(startTime).Seconds()

	// Display final message
	fmt.Print("\033[2J\033[H")
	if g.win {
		fmt.Println("YOU WIN!")
	} else {
		fmt.Println("GAME OVER!")
	}
	fmt.Printf("Final Score: %d\n", g.score)

	// Save score if logged in
	if authManager != nil {
		if realAuth, ok := authManager.(*auth.CLIAuth); ok && realAuth.GetSession().IsLoggedIn() {
			session := realAuth.GetSession().GetCurrentSession()
			if realDB, ok := db.(*database.DB); ok && session != nil {
				metadata := map[string]interface{}{"duration": duration, "won": g.win}
				realDB.SaveGameScore(session.UserID, "breakout", g.score, metadata)
			}
		}
	}

	return &types.GameResult{
		GameName: "Breakout",
		Score:    g.score,
		Duration: duration,
		Metadata: map[string]interface{}{"won": g.win},
	}
}

// inputHandler runs in a goroutine to listen for keyboard events
func (g *Game) inputHandler() {
	for {
		// Capture the character (rune) as well as the key
		char, key, err := keyboard.GetKey()
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Check for 'q' (case-insensitive) or the Esc key to quit
		if char == 'q' || char == 'Q' || key == keyboard.KeyEsc {
			close(g.inputChan)
			return
		}

		// Send the appropriate input to the processing channel.
		// If a special key (like an arrow key) is pressed, send it.
		// Otherwise, send the character rune as a Key type.
		if key != 0 {
			g.inputChan <- key
		} else if char != 0 {
			g.inputChan <- keyboard.Key(char)
		}
	}
}

// processInput handles the logic for a key press
func (g *Game) processInput() {
	select {
	case key, ok := <-g.inputChan:
		if !ok {
			g.gameOver = true
			return
		}
		switch key {
		// Check for both character and arrow keys for movement
		case keyboard.KeyArrowLeft, 'a', 'A':
			if g.paddle.X > 0 {
				g.paddle.X -= 2 // Move faster
			}
		case keyboard.KeyArrowRight, 'd', 'D':
			if g.paddle.X+g.paddle.Size < BoardWidth {
				g.paddle.X += 2 // Move faster
			}
		}
	default:
		// No input, do nothing
	}
}

// nextLevel prepares the game for the next stage.
func (g *Game) nextLevel() {
	g.level++
	g.score += 100 * (g.level - 1) // Bonus points for clearing a level

	// Reset ball to center and increase its speed by 10%
	g.ball.X = float64(BoardWidth / 2)
	g.ball.Y = float64(BoardHeight / 2)
	g.ball.VX *= 1.1
	g.ball.VY *= 1.1

	g.resetBricks()

	fmt.Printf("LEVEL %d\n", g.level)
	time.Sleep(2 * time.Second)
}

func (g *Game) update() {
	// Move ball
	g.ball.X += g.ball.VX
	g.ball.Y += g.ball.VY

	// Wall collision (left/right)
	if g.ball.X <= 0 || g.ball.X >= float64(BoardWidth-1) {
		g.ball.VX = -g.ball.VX
	}

	// Wall collision (top)
	if g.ball.Y <= 0 {
		g.ball.VY = -g.ball.VY
	}

	// Paddle collision
	ballX, ballY := int(g.ball.X), int(g.ball.Y)
	if ballY == g.paddle.Y && ballX >= g.paddle.X && ballX <= g.paddle.X+g.paddle.Size {
		g.ball.VY = -g.ball.VY
	}

	// Brick collision
	bricksLeft := false
	for i := range g.bricks {
		if g.bricks[i].Alive {
			bricksLeft = true
			if ballY == g.bricks[i].Y && ballX >= g.bricks[i].X && ballX < g.bricks[i].X+3 {
				g.bricks[i].Alive = false
				g.ball.VY = -g.ball.VY
				g.score += g.config.BrickScore // Use config for score
			}
		}
	}
	if !bricksLeft {
		g.nextLevel() // Call nextLevel instead of ending the game
		return        // Skip the rest of the update for this frame
	}

	// Ball lost
	if g.ball.Y >= float64(BoardHeight-1) {
		g.lives--
		if g.lives <= 0 {
			g.gameOver = true
		} else {
			// Reset ball
			g.ball.X = float64(BoardWidth / 2)
			g.ball.Y = float64(BoardHeight / 2)
		}
	}
}

func (g *Game) render() {
	fmt.Print("\033[2J\033[H") // Clear screen

	// Top border
	fmt.Println("╔" + strings.Repeat("═", BoardWidth) + "╗")

	// Game area
	for y := 0; y < BoardHeight; y++ {
		fmt.Print("║")
		for x := 0; x < BoardWidth; x++ {
			// Draw ball
			if int(g.ball.X) == x && int(g.ball.Y) == y {
				fmt.Print("o")
				continue
			}

			// Draw paddle
			if y == g.paddle.Y && x >= g.paddle.X && x < g.paddle.X+g.paddle.Size {
				fmt.Print("=")
				continue
			}

			// Draw bricks
			isBrick := false
			for _, brick := range g.bricks {
				if brick.Alive && y == brick.Y && x >= brick.X && x < brick.X+3 {
					fmt.Print("▇")
					isBrick = true
					break
				}
			}
			if isBrick {
				continue
			}

			fmt.Print(" ")
		}
		fmt.Println("║")
	}

	// Bottom border
	fmt.Println("╚" + strings.Repeat("═", BoardWidth) + "╝")
	fmt.Printf("Score: %d | Lives: %d | Level: %d\n", g.score, g.lives, g.level)
}

// loadConfig reads settings from breakout.lua
func loadConfig() *GameConfig {
	// Default config in case Lua file is missing or fails
	defaultLayout := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 0, 0, 0, 0, 0, 0, 1, 1},
		{1, 1, 0, 0, 0, 0, 0, 0, 1, 1},
	}
	defaultConfig := &GameConfig{
		PaddleSize:  8,
		BallSpeedX:  0.5,
		BallSpeedY:  -0.25,
		Lives:       3,
		BrickScore:  10,
		BrickLayout: defaultLayout,
	}

	if _, err := os.Stat("games/breakout/breakout.lua"); os.IsNotExist(err) {
		log.Println("breakout.lua not found, using default config")
		return defaultConfig
	}

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile("games/breakout/breakout.lua"); err != nil {
		log.Printf("Error loading breakout.lua: %v. Using default config.", err)
		return defaultConfig
	}

	// Get the returned value from the script, which is at the top of the stack.
	breakoutTable := L.Get(-1)
	if tbl, ok := breakoutTable.(*lua.LTable); ok {
		configTable := tbl.RawGetString("config")
		layoutsTable := tbl.RawGetString("layouts") // Get layouts table
		standardLayout := defaultLayout

		// Safely parse the layout
		if layoutsTbl, ok := layoutsTable.(*lua.LTable); ok {
			standardLayoutTable := layoutsTbl.RawGetString("standard")
			if layout, ok := standardLayoutTable.(*lua.LTable); ok {
				parsedLayout := [][]int{}
				layout.ForEach(func(_, rowVal lua.LValue) {
					if rowTbl, ok := rowVal.(*lua.LTable); ok {
						parsedRow := []int{}
						rowTbl.ForEach(func(_, cellVal lua.LValue) {
							if cellNum, ok := cellVal.(lua.LNumber); ok {
								parsedRow = append(parsedRow, int(cellNum))
							}
						})
						parsedLayout = append(parsedLayout, parsedRow)
					}
				})
				if len(parsedLayout) > 0 {
					standardLayout = parsedLayout
				}
			}
		}

		if configTbl, ok := configTable.(*lua.LTable); ok {
			return &GameConfig{
				PaddleSize:  getLuaInt(configTbl, "paddle_size", defaultConfig.PaddleSize),
				BallSpeedX:  getLuaFloat(configTbl, "ball_speed_x", defaultConfig.BallSpeedX),
				BallSpeedY:  getLuaFloat(configTbl, "ball_speed_y", defaultConfig.BallSpeedY),
				Lives:       getLuaInt(configTbl, "lives", defaultConfig.Lives),
				BrickScore:  getLuaInt(configTbl, "brick_score", defaultConfig.BrickScore),
				BrickLayout: standardLayout, // Assign the loaded layout
			}
		}
	}

	log.Println("Could not parse breakout.config table, using default config")
	return defaultConfig
}

// Helper functions to safely get values from a Lua table
func getLuaInt(tbl *lua.LTable, key string, fallback int) int {
	val := tbl.RawGetString(key)
	if num, ok := val.(lua.LNumber); ok {
		return int(num)
	}
	return fallback
}

func getLuaFloat(tbl *lua.LTable, key string, fallback float64) float64 {
	val := tbl.RawGetString(key)
	if num, ok := val.(lua.LNumber); ok {
		return float64(num)
	}
	return fallback
}
