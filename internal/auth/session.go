package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Session represents a user session
type Session struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// SessionManager handles user sessions
type SessionManager struct {
	sessionFile string
	current     *Session
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	// Try to load existing session and log if it fails
	if err := sm.LoadSession(); err != nil {
		// This is not a fatal error, just means no session was loaded.
		// A log message is useful for debugging.
		log.Printf("[DEBUG] No previous session found or failed to load: %v", err)
	}
	return sm
}

// SaveSession saves the current session to disk
func (sm *SessionManager) SaveSession(userID int, username, email string) error {
	sm.current = &Session{
		UserID:   userID,
		Username: username,
		Email:    email,
	}

	data, err := json.Marshal(sm.current)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	err = os.WriteFile(sm.sessionFile, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// LoadSession loads a session from disk
func (sm *SessionManager) LoadSession() error {
	data, err := os.ReadFile(sm.sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No session file exists, which is fine
		}
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	sm.current = &session
	return nil
}

// GetCurrentSession returns the current session data
func (sm *SessionManager) GetCurrentSession() *Session {
	return sm.current
}

// IsLoggedIn returns true if a user is currently logged in
func (sm *SessionManager) IsLoggedIn() bool {
	return sm.current != nil
}

// ClearSession clears the current session
func (sm *SessionManager) ClearSession() error {
	sm.current = nil

	if _, err := os.Stat(sm.sessionFile); err == nil {
		err = os.Remove(sm.sessionFile)
		if err != nil {
			return fmt.Errorf("failed to remove session file: %w", err)
		}
	}

	return nil
}

// GetUserInfo returns formatted user information
func (sm *SessionManager) GetUserInfo() string {
	if sm.current == nil {
		return "Not logged in"
	}
	return fmt.Sprintf("Logged in as: %s (%s)", sm.current.Username, sm.current.Email)
}
