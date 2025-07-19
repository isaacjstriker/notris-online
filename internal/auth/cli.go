package auth

import (
	"fmt"

	"github.com/isaacjstriker/devware/internal/database"
	"github.com/isaacjstriker/devware/ui"
)

// CLIAuth handles authentication through the CLI
type CLIAuth struct {
	db      *database.DB
	session *SessionManager
}

// NewCLIAuth creates a new CLI authentication handler
func NewCLIAuth(db *database.DB) *CLIAuth {
	return &CLIAuth{
		db:      db,
		session: NewSessionManager(),
	}
}

// GetSession returns the current session manager
func (auth *CLIAuth) GetSession() *SessionManager {
	return auth.session
}

// ShowAuthMenu displays the authentication menu
func (auth *CLIAuth) ShowAuthMenu() {
	for {
		var menuItems []ui.MenuItem

		if auth.session.IsLoggedIn() {
			// User is logged in
			menuItems = []ui.MenuItem{
				{Label: fmt.Sprintf("👤 Currently: %s", auth.session.GetUserInfo()), Value: "info"},
				{Label: "🔄 Switch Account", Value: "switch"},
				{Label: "🚪 Logout", Value: "logout"},
				{Label: "⬅️  Back to Main Menu", Value: "back"},
			}
		} else {
			// User is not logged in
			menuItems = []ui.MenuItem{
				{Label: "🔑 Login", Value: "login"},
				{Label: "📝 Register New Account", Value: "register"},
				{Label: "👻 Continue as Guest", Value: "guest"},
				{Label: "⬅️  Back to Main Menu", Value: "back"},
			}
		}

		menu := ui.NewMenu("Authentication", menuItems)
		choice := menu.Show()

		switch choice {
		case "login":
			auth.handleLogin()
		case "register":
			auth.handleRegister()
		case "guest":
			fmt.Println("\n👻 Continuing as guest...")
			fmt.Println("Note: Your scores won't be saved!")
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
			return
		case "switch":
			auth.session.ClearSession()
			fmt.Println("\n🔄 Logged out. Please login with a different account.")
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
		case "logout":
			auth.handleLogout()
		case "info":
			fmt.Printf("\n%s\n", auth.session.GetUserInfo())
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
		case "back", "exit":
			return
		}
	}
}

// handleLogin handles user login
func (auth *CLIAuth) handleLogin() {
	fmt.Println("\n🔑 Login to Your Account")
	fmt.Println("========================")

	username, err := ReadInput("Username: ")
	if err != nil {
		fmt.Printf("Error reading username: %v\n", err)
		return
	}

	password, err := ReadPassword("Password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}

	user, passwordHash, err := auth.db.GetUserByUsername(username)
	if err != nil {
		fmt.Println("❌ Invalid username or password")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	if !CheckPassword(password, passwordHash) {
		fmt.Println("❌ Invalid username or password")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	err = auth.session.SaveSession(user.ID, user.Username, user.Email)
	if err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		return
	}

	fmt.Printf("✅ Welcome back, %s!\n", user.Username)
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

// handleRegister handles user registration
func (auth *CLIAuth) handleRegister() {
	fmt.Println("\n📝 Create New Account")
	fmt.Println("=====================")

	username, err := ReadInput("Username (3-50 characters): ")
	if err != nil {
		fmt.Printf("Error reading username: %v\n", err)
		return
	}

	if err := ValidateUsername(username); err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	email, err := ReadInput("Email: ")
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		return
	}

	if err := ValidateEmail(email); err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	password, err := ReadPassword("Password (8+ characters): ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}

	if err := ValidatePassword(password); err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	confirmPassword, err := ReadPassword("Confirm Password: ")
	if err != nil {
		fmt.Printf("Error reading confirmation: %v\n", err)
		return
	}

	if password != confirmPassword {
		fmt.Println("❌ Passwords do not match")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		return
	}

	user, err := auth.db.CreateUser(username, email, passwordHash)
	if err != nil {
		fmt.Printf("❌ Failed to create account: %v\n", err)
		fmt.Println("(Username or email might already be taken)")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	err = auth.session.SaveSession(user.ID, user.Username, user.Email)
	if err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		return
	}

	fmt.Printf("✅ Account created successfully! Welcome, %s!\n", user.Username)
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

// handleLogout handles user logout
func (auth *CLIAuth) handleLogout() {
	var username string
	if auth.session.IsLoggedIn() {
		session := auth.session.GetCurrentSession()
		if session != nil {
			username = session.Username
		}
	}

	err := auth.session.ClearSession()
    if err != nil {
        fmt.Printf("Error clearing session: %v\n", err)
        return
    }

    if username != "" {
        fmt.Printf("👋 Goodbye, %s! You have been logged out.\n", username)
    } else {
        fmt.Println("👋 You have been logged out.")
    }
    fmt.Println("Press Enter to continue...")
    fmt.Scanln()
}

// RequireAuth ensures user is authenticated, prompting login if needed
func (auth *CLIAuth) RequireAuth() bool {
	if auth.session.IsLoggedIn() {
		return true
	}

	fmt.Println("\n🔒 Authentication Required")
	fmt.Println("You need to be logged in to save your scores!")
	fmt.Println()

	continueItems := []ui.MenuItem{
		{Label: "🔑 Login Now", Value: "login"},
		{Label: "📝 Create Account", Value: "register"},
		{Label: "👻 Continue as Guest (no scores saved)", Value: "guest"},
	}

	menu := ui.NewMenu("Authentication Required", continueItems)
	choice := menu.Show()

	switch choice {
	case "login":
		auth.handleLogin()
		return auth.session.IsLoggedIn()
	case "register":
		auth.handleRegister()
		return auth.session.IsLoggedIn()
	case "guest":
		return false
	}

	return false
}
