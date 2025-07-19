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

func (m *Menu) calculateWidth() int {
	maxWidth := len(m.Title) + 6 // Title + some padding

	// Check each menu item
	for _, item := range m.Items {
		// Account for selection indicators and padding
		itemWidth := len(item.Label) + 10 // "► " + label + " ◄" + padding
		if itemWidth > maxWidth {
			maxWidth = itemWidth
		}
	}

	// Ensure reasonable bounds
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 70 {
		maxWidth = 70
	}

	// Make it even for better centering
	if maxWidth%2 != 0 {
		maxWidth++
	}

	return maxWidth
}

func (m *Menu) render() {
	// Clear screen
	fmt.Print("\033[H\033[2J")

	// Calculate width
	m.Width = m.calculateWidth()

	// ASCII Art Title (centered)
	fmt.Println(`
    ██████╗ ███████╗██╗   ██╗██╗    ██╗ █████╗ ██████╗ ███████╗
    ██╔══██╗██╔════╝██║   ██║██║    ██║██╔══██╗██╔══██╗██╔════╝
    ██║  ██║█████╗  ██║   ██║██║ █╗ ██║███████║██████╔╝█████╗  
    ██║  ██║██╔══╝  ╚██╗ ██╔╝██║███╗██║██╔══██║██╔══██╗██╔══╝  
    ██████╔╝███████╗ ╚████╔╝ ╚███╔███╔╝██║  ██║██║  ██║███████╗
    ╚═════╝ ╚══════╝  ╚═══╝   ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
    `)

	// Calculate menu centering
	logoWidth := 67 // Approximate width of the ASCII art
	menuIndent := (logoWidth - m.Width) / 2
	if menuIndent < 0 {
		menuIndent = 0
	}

	// Simple, clean border
	borderChar := "="
	sideChar := "|"

	// Top border (centered)
	fmt.Printf("%s+%s+\n", strings.Repeat(" ", menuIndent), strings.Repeat(borderChar, m.Width-2))

	// Title (centered)
	titlePadding := (m.Width - len(m.Title) - 2) / 2
	titleRemainder := m.Width - len(m.Title) - titlePadding - 2
	titleLine := strings.Repeat(" ", titlePadding) + m.Title + strings.Repeat(" ", titleRemainder)
	fmt.Printf("%s%s%s%s\n", strings.Repeat(" ", menuIndent), sideChar, titleLine, sideChar)

	// Separator (centered)
	fmt.Printf("%s+%s+\n", strings.Repeat(" ", menuIndent), strings.Repeat(borderChar, m.Width-2))

	// Menu items (centered)
	for i, item := range m.Items {
		prefix := "  "
		suffix := "  "

		if i == m.Selected {
			prefix = "> "
			suffix = " <"
		}

		itemText := prefix + item.Label + suffix
		padding := m.Width - len(itemText) - 2

		if padding < 0 {
			// Truncate if too long
			maxLen := m.Width - 8 // Account for prefix, suffix, borders
			if maxLen > 0 {
				truncated := item.Label
				if len(truncated) > maxLen {
					truncated = truncated[:maxLen-3] + "..."
				}
				itemText = prefix + truncated + suffix
				padding = m.Width - len(itemText) - 2
			}
		}

		// Ensure padding is not negative
		if padding < 0 {
			padding = 0
		}

		// Create the final item line with proper padding
		itemLine := itemText + strings.Repeat(" ", padding)

		// Print the line with highlighting if selected (centered)
		if i == m.Selected {
			fmt.Printf("%s%s\033[7m%s\033[0m%s\n", strings.Repeat(" ", menuIndent), sideChar, itemLine, sideChar)
		} else {
			fmt.Printf("%s%s%s%s\n", strings.Repeat(" ", menuIndent), sideChar, itemLine, sideChar)
		}
	}

	// Bottom border (centered)
	fmt.Printf("%s+%s+\n", strings.Repeat(" ", menuIndent), strings.Repeat(borderChar, m.Width-2))

	fmt.Println()
	fmt.Printf("%sUse ↑/↓ arrows to navigate, Enter to select, 'q' to quit\n", strings.Repeat(" ", menuIndent))
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
