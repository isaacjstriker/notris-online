package typing

import (
	"os"
	"time"
	"bufio"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func Run() {
	L := lua.NewState()
	defer L.Close()

	// Load Lua script
	if err := L.DoFile("games/typing_game.lua"); err != nil {
		fmt.Println("Lua error:", err)
	}
	
	// Get words table from Lua
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("get_words"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		fmt.Println("Lua error:", err)
		return
	}
	wordsTable := L.Get(-1)
	L.Pop(1)

	words := []string{}
	if tbl, ok := wordsTable.(*lua.LTable); ok {
		tbl.ForEach(func(_, value lua.LValue) {
			words = append(words, value.String())
		})
	}

	timer := time.Now()

	// Read user input with timeout
	for _, word := range words {
		fmt.Printf("You have 5 seconds to type this word: %s\n", word)
		inputCh := make(chan string)
		go func() {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			inputCh <- text[:len(text)-1] // <- Remove newline
		}()

		success := false
		select {
		case input := <-inputCh:
			// Call Lua function to check the word
			L.Push(L.GetGlobal("check_word"))
			L.Push(lua.LString(input))
			L.Push(lua.LString(word))
			if err := L.PCall(2, 1, nil); err != nil {
				fmt.Println("Lua error:", err)
				return
			}
			result := L.Get(-1).(lua.LBool)
			L.Pop(1)
			if result {
				fmt.Println("Correct!")
				success = true
			} else {
				fmt.Println("Incorrect. Game over.")
			}
		case <-time.After(5 * time.Second):
			fmt.Println("\nTime's up! Game over.")
		}
		if !success {
			break
		}
	}
	elapsed := time.Since(timer)
	fmt.Printf("Closing... You took %.2f seconds\n", elapsed.Seconds())
}