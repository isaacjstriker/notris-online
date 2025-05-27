package typing

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	lua "github.com/yuin/gopher-lua"
)

func Run() {
	L := lua.NewState()
	defer L.Close()

	// Load Lua script
	if err := L.DoFile("games/typing_game.lua"); err != nil {
		fmt.Println("Lua error:", err)
		return
	}

	// Get random words list
	L.Push(L.GetGlobal("get_random_words"))
	L.Push(lua.LNumber(10)) // Change number to change amount of words cycled
	if err := L.PCall(1, 1, nil); err != nil {
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

	// Shuffle the words slice
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	timer := time.Now()

	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to open keyboard:", err)
		return
	}
	defer keyboard.Close()

	// Read user input with timeout
	for _, word := range words {
		fmt.Printf("You have 5 seconds to type this word: %s\n", word)
		inputCh := make(chan string)
		go func() {
			input := ""
			for {
				char, key, err := keyboard.GetKey()
				if err != nil {
					inputCh <- ""
					return
				}
				if key == keyboard.KeyEnter {
					inputCh <- input
					return
				} else if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
					if len(input) > 0 {
						input = input[:len(input)-1]
						fmt.Print("\b \b")
					}
				} else if key == 0 {
					input += string(char)
					fmt.Print(string(char))
				}
			}
		}()

		select {
		case input := <-inputCh:
			fmt.Println()
			// Call Lua function to check the word
			L.Push(L.GetGlobal("check_word"))
			L.Push(lua.LString(strings.TrimSpace(input)))
			L.Push(lua.LString(word))
			if err := L.PCall(2, 1, nil); err != nil {
				fmt.Println("Lua error:", err)
				return
			}
			result := L.Get(-1)
			L.Pop(1)
			if result == lua.LTrue {
				fmt.Println("Correct!")
			} else {
				fmt.Println("Incorrect. Game over.")
				return
			}
		case <-time.After(5 * time.Second):
			fmt.Println("\nTime's up! Game over.")
			return
		}
	}
	elapsed := time.Since(timer)
	fmt.Printf("Closing... You took %.2f seconds\n", elapsed.Seconds())
}
