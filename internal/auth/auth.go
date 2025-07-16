package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ReadPassword reads a password from stdin without echoing it
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // Add a newline after password input
	return string(bytePassword), nil
}

// ReadInput reads a line of input from stdin
func ReadInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must be no more than 50 characters long")
	}
	// Add more validation rules as needed
	return nil
}

// ValidateEmail checks if an email is valid (basic validation)
func ValidateEmail(email string) error {
	if len(email) < 5 {
		return fmt.Errorf("email must be at least 5 characters long")
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("email must contain @ symbol")
	}
	if !strings.Contains(email, ".") {
		return fmt.Errorf("email must contain a domain")
	}
	return nil
}

// ValidatePassword checks if a password meets requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	// Add more validation rules as needed (uppercase, numbers, symbols)
	return nil
}
