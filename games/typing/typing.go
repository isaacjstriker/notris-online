package typing

import (
	"fmt"
	"time"

	"github.com/isaacjstriker/devware/internal/auth"
	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/internal/types"
	lua "github.com/yuin/gopher-lua"
)

// TypingGame implements the Game interface
type TypingGame struct {
	luaScriptPath string
}

// NewTypingGame creates a new typing game instance
func NewTypingGame() *TypingGame {
	return &TypingGame{
		luaScriptPath: "games/typing/typing_game.lua",
	}
}

func (tg *TypingGame) GetName() string {
	return "Typing Speed Challenge"
}

func (tg *TypingGame) GetDescription() string {
	return "Test your typing speed and accuracy with random words"
}

func (tg *TypingGame) GetDifficulty() int {
	return 3 // Medium difficulty
}

func (tg *TypingGame) IsAvailable() bool {
	// Check if Lua file exists
	L := lua.NewState()
	defer L.Close()

	err := L.DoFile(tg.luaScriptPath)
	return err == nil
}

func (tg *TypingGame) Play(db interface{}, authManager interface{}) *types.GameResult {
	// Type assert the interfaces to the specific types you need
	var realDB *database.DB
	var realAuth *auth.CLIAuth

	if db != nil {
		var ok bool
		realDB, ok = db.(*database.DB)
		if !ok {
			return &types.GameResult{
				GameName: tg.GetName(),
				Score:    0,
				Duration: 0,
				Accuracy: 0,
				Perfect:  false,
			}
		}
	}

	if authManager != nil {
		var ok bool
		realAuth, ok = authManager.(*auth.CLIAuth)
		if !ok {
			return &types.GameResult{
				GameName: tg.GetName(),
				Score:    0,
				Duration: 0,
				Accuracy: 0,
				Perfect:  false,
			}
		}
	}

	// Now use realDB and realAuth in your existing logic
	stats, err := tg.playGame()
	if err != nil {
		fmt.Printf("[ERROR] Game error: %v\n", err)
		return &types.GameResult{
			GameName: tg.GetName(),
			Score:    0,
			Duration: 0,
			Accuracy: 0,
			Perfect:  false,
		}
	}

	// Convert to GameResult format
	result := &types.GameResult{
		GameName: tg.GetName(),
		Score:    stats.Score,
		Duration: stats.TotalTime,
		Accuracy: stats.Accuracy,
		Perfect:  stats.Accuracy >= 95.0,
		Metadata: map[string]interface{}{
			"wpm":           stats.WPM,
			"words_typed":   stats.WordsTyped,
			"correct_words": stats.CorrectWords,
		},
	}

	// Calculate bonus for perfect games
	if result.Perfect {
		result.Bonus = int(float64(result.Score) * 0.2) // 20% bonus
		result.Score += result.Bonus
	}

	// Save to database if user is authenticated
	if realDB != nil && realAuth != nil && realAuth.GetSession().IsLoggedIn() {
		session := realAuth.GetSession().GetCurrentSession()
		if session != nil {
			additionalData := map[string]interface{}{
				"wpm":           stats.WPM,
				"accuracy":      stats.Accuracy,
				"words_typed":   stats.WordsTyped,
				"correct_words": stats.CorrectWords,
				"duration":      stats.TotalTime,
				"perfect":       result.Perfect,
				"bonus":         result.Bonus,
			}

			err := realDB.SaveGameScore(session.UserID, "typing", result.Score, additionalData)
			if err != nil {
				fmt.Printf("[WARNING] Failed to save score: %v\n", err)
			} else {
				fmt.Println("[OK] Score saved to database!")
			}
		}
	}

	fmt.Printf("\n[SCORE] Final Score: %d\n", result.Score)

	// Check if user is logged in
	if realAuth == nil || !realAuth.GetSession().IsLoggedIn() {
		fmt.Println("[WARNING] Not logged in - score won't be saved!")
		fmt.Println("[INFO] Log in to save your scores to the leaderboard.")
		return result
	}

	session := realAuth.GetSession().GetCurrentSession()
	if session == nil {
		fmt.Println("[ERROR] No valid session - score won't be saved!")
		return result
	}

	fmt.Printf("[SAVE] Attempting to save score for user: %s (ID: %d)\n", session.Username, session.UserID)

	// Submit score to database
	err = realDB.SubmitScore(session.UserID, "typing", result.Score, nil)
	if err != nil {
		fmt.Printf("[ERROR] Error saving score: %v\n", err)
		return result
	}

	fmt.Println("[OK] Score saved successfully!")

	return result
}

// playGame runs the Lua script and returns the game statistics
func (tg *TypingGame) playGame() (*types.GameStats, error) {
	L := lua.NewState()
	defer L.Close()

	// Register Go functions that Lua can call
	tg.registerGoFunctions(L)

	// Load and execute the Lua script
	if err := L.DoFile(tg.luaScriptPath); err != nil {
		return nil, fmt.Errorf("failed to load Lua script: %w", err)
	}

	// Call the main game function in Lua
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("run_typing_game"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil, fmt.Errorf("failed to run Lua game function: %w", err)
	}

	// Get the result from Lua
	result := L.Get(-1)
	L.Pop(1)

	// Convert Lua table to Go struct
	stats, err := tg.luaTableToGameStats(L, result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Lua result: %w", err)
	}

	return stats, nil
}

// registerGoFunctions registers Go functions that Lua can call
func (tg *TypingGame) registerGoFunctions(L *lua.LState) {
	// Function for Lua to print messages
	L.SetGlobal("go_print", L.NewFunction(func(L *lua.LState) int {
		message := L.ToString(1)
		fmt.Print(message)
		return 0
	}))

	// Function for Lua to print lines
	L.SetGlobal("go_println", L.NewFunction(func(L *lua.LState) int {
		message := L.ToString(1)
		fmt.Println(message)
		return 0
	}))

	// Function for Lua to read user input
	L.SetGlobal("go_read_line", L.NewFunction(func(L *lua.LState) int {
		var input string
		fmt.Scanln(&input)
		L.Push(lua.LString(input))
		return 1
	}))

	// Function for Lua to get current time
	L.SetGlobal("go_current_time", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(float64(time.Now().UnixNano()) / 1e9))
		return 1
	}))

	// Function for Lua to sleep
	L.SetGlobal("go_sleep", L.NewFunction(func(L *lua.LState) int {
		seconds := L.ToNumber(1)
		time.Sleep(time.Duration(float64(seconds) * float64(time.Second)))
		return 0
	}))
}

// luaTableToGameStats converts a Lua table to a GameStats struct
func (tg *TypingGame) luaTableToGameStats(_ *lua.LState, lv lua.LValue) (*types.GameStats, error) {
	table, ok := lv.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("expected Lua table, got %T", lv)
	}

	stats := &types.GameStats{}

	// Extract values from Lua table
	if score := table.RawGetString("score"); score != lua.LNil {
		if num, ok := score.(lua.LNumber); ok {
			stats.Score = int(num)
		}
	}

	if wpm := table.RawGetString("wpm"); wpm != lua.LNil {
		if num, ok := wpm.(lua.LNumber); ok {
			stats.WPM = float64(num)
		}
	}

	if accuracy := table.RawGetString("accuracy"); accuracy != lua.LNil {
		if num, ok := accuracy.(lua.LNumber); ok {
			stats.Accuracy = float64(num)
		}
	}

	if wordsTyped := table.RawGetString("words_typed"); wordsTyped != lua.LNil {
		if num, ok := wordsTyped.(lua.LNumber); ok {
			stats.WordsTyped = int(num)
		}
	}

	if correctWords := table.RawGetString("correct_words"); correctWords != lua.LNil {
		if num, ok := correctWords.(lua.LNumber); ok {
			stats.CorrectWords = int(num)
		}
	}

	if totalTime := table.RawGetString("total_time"); totalTime != lua.LNil {
		if num, ok := totalTime.(lua.LNumber); ok {
			stats.TotalTime = float64(num)
		}
	}

	return stats, nil
}
