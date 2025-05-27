package main

import (
	"bufio"
	"strings"
	"fmt"
	"os"
	"github.com/isaacjstriker/devware/games/typing"
)

func main() {
	for {
		fmt.Println("Welcome to Dev Ware!")
		fmt.Println("Select a game:")
		fmt.Println("1) Typing Game")
		fmt.Print("Enter the number of your choice: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			typing.Run()
		case "0", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Please make a valid selection.")
		}
	}
}