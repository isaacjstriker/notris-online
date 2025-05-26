package main

import (
	"fmt"
	"os"
	"github.com/isaacjstriker/devware/games/typing"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: devware <game>")
		return
	}
	switch os.Args[1] {
	case "typing":
		typing.Run()
	default:
		fmt.Println("Unknown game:", os.Args[1])
	}
}