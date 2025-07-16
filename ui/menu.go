package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/eiannone/keyboard"
)

type MenuItem struct {
	Label string
	Value string
}

type Menu struct {
	Title    string
	Items    []MenuItem
	Selected int
	Width    int
}

func NewMenu(title string, items []MenuItem) *Menu {
	return &Menu{
		Title:    title,
		Items:    items,
		Selected: 0,
		Width:    60,
	}
}

func (m *Menu) clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (m *Menu) drawBorder(char string, length int) {
	fmt.Print("╔")
	for i := 0; i < length-2; i++ {
		fmt.Print("═")
	}
	fmt.Println("╗")
}

func (m *Menu) drawBottomBorder(length int) {
	fmt.Print("╚")
	for i := 0; i < length-2; i++ {
		fmt.Print("═")
	}
	fmt.Println("╝")
}

func (m *Menu) centerText(text string, width int) string {
	if len(text) >= width-4 {
		return text[:width-4]
	}
	padding := (width - len(text) - 4) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding-4)
}

func (m *Menu) render() {
	m.clearScreen()

	// ASCII Art Title
	fmt.Print(`
██████╗ ███████╗██╗   ██╗    ██╗    ██╗ █████╗ ██████╗ ███████╗
██╔══██╗██╔════╝██║   ██║    ██║    ██║██╔══██╗██╔══██╗██╔════╝
██║  ██║█████╗  ██║   ██║    ██║ █╗ ██║███████║██████╔╝█████╗  
██║  ██║██╔══╝  ╚██╗ ██╔╝    ██║███╗██║██╔══██║██╔══██╗██╔══╝  
██████╔╝███████╗ ╚████╔╝     ╚███╔███╔╝██║  ██║██║  ██║███████╗
╚═════╝ ╚══════╝  ╚═══╝       ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
`)
	fmt.Println()

	fmt.Println(m.centerText("🎮 Professional Game Collection 🎮", m.Width))
	fmt.Println()

	// Draw menu border
	m.drawBorder("═", m.Width)

	// Draw title
	titleText := m.centerText(m.Title, m.Width)
	fmt.Printf("║%s║\n", titleText)

	// Draw separator
	fmt.Print("╠")
	for i := 0; i < m.Width-2; i++ {
		fmt.Print("═")
	}
	fmt.Println("╣")

	// Draw menu items
	for i, item := range m.Items {
		var prefix string
		if i == m.Selected {
			prefix = "► "
		} else {
			prefix = "  "
		}

		itemText := prefix + item.Label
		paddedText := m.centerText(itemText, m.Width)

		if i == m.Selected {
			fmt.Printf("║\033[7m%s\033[0m║\n", paddedText) // Highlighted
		} else {
			fmt.Printf("║%s║\n", paddedText)
		}
	}

	// Draw bottom border
	m.drawBottomBorder(m.Width)

	fmt.Println()
	fmt.Println("Use ↑/↓ arrows to navigate, Enter to select, 'q' to quit")
}

func (m *Menu) moveUp() {
	if m.Selected > 0 {
		m.Selected--
	} else {
		m.Selected = len(m.Items) - 1 // Wrap to bottom
	}
}

func (m *Menu) moveDown() {
	if m.Selected < len(m.Items)-1 {
		m.Selected++
	} else {
		m.Selected = 0 // Wrap to top
	}
}

func (m *Menu) Show() string {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("Failed to open keyboard: %v\n", err)
		return ""
	}
	defer keyboard.Close()

	for {
		m.render()

		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading key: %v\n", err)
			return ""
		}

		switch key {
		case keyboard.KeyArrowUp:
			m.moveUp()
		case keyboard.KeyArrowDown:
			m.moveDown()
		case keyboard.KeyEnter:
			return m.Items[m.Selected].Value
		case keyboard.KeyEsc:
			return "exit"
		}

		if char == 'q' || char == 'Q' {
			return "exit"
		}
	}
}
