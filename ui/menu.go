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
	// The loop was correct, but the final character needs to be printed outside.
	fmt.Print(strings.Repeat("═", length-2))
	fmt.Println("╝")
}

// centerText now correctly centers the text within the available content width.
func (m *Menu) centerText(text string, contentWidth int) string {
	if len(text) >= contentWidth {
		// Truncate if the text is too long to fit
		return text[:contentWidth]
	}
	padding := contentWidth - len(text)
	leftPadding := padding / 2
	rightPadding := padding - leftPadding
	return strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
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
	// Use your clearScreen function instead of escape codes
	m.clearScreen()

	// Calculate width
	m.Width = m.calculateWidth()

	// ASCII Art Title
	fmt.Println(`
    ██████╗ ███████╗██╗   ██╗██╗    ██╗ █████╗ ██████╗ ███████╗
    ██╔══██╗██╔════╝██║   ██║██║    ██║██╔══██╗██╔══██╗██╔════╝
    ██║  ██║█████╗  ██║   ██║██║ █╗ ██║███████║██████╔╝█████╗  
    ██║  ██║██╔══╝  ╚██╗ ██╔╝██║███╗██║██╔══██║██╔══██╗██╔══╝  
    ██████╔╝███████╗ ╚████╔╝ ╚███╔███╔╝██║  ██║██║  ██║███████╗
    ╚═════╝ ╚══════╝  ╚═══╝   ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
    `)

	// Calculate menu centering
	logoWidth := 67
	menuIndent := (logoWidth - m.Width) / 2
	if menuIndent < 0 {
		menuIndent = 0
	}

	// Use your drawBorder function with proper indentation
	fmt.Print(strings.Repeat(" ", menuIndent))
	m.drawBorder("═", m.Width)

	// The content width is the total width minus the two side borders
	contentWidth := m.Width - 2

	// Use the corrected centerText function for the title
	centeredTitle := m.centerText(m.Title, contentWidth)
	fmt.Printf("%s║%s║\n", strings.Repeat(" ", menuIndent), centeredTitle)

	// Separator line
	fmt.Printf("%s╠%s╣\n", strings.Repeat(" ", menuIndent), strings.Repeat("═", contentWidth))

	// Menu items
	for i, item := range m.Items {
		prefix := "  "
		suffix := "  "

		if i == m.Selected {
			prefix = "► "
			suffix = " ◄"
		}

		itemText := prefix + item.Label
		// Calculate padding to make the item fill the width
		padding := contentWidth - len(itemText) - len(suffix)
		if padding < 0 {
			padding = 0
		}
		fullItemText := itemText + strings.Repeat(" ", padding) + suffix

		// Print with highlighting if selected
		if i == m.Selected {
			// Use inverse video for highlighting
			fmt.Printf("%s║\033[7m%s\033[0m║\n", strings.Repeat(" ", menuIndent), fullItemText)
		} else {
			fmt.Printf("%s║%s║\n", strings.Repeat(" ", menuIndent), fullItemText)
		}
	}

	// Use your drawBottomBorder function
	fmt.Print(strings.Repeat(" ", menuIndent))
	m.drawBottomBorder(m.Width)

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
