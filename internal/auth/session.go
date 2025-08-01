package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Session struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SessionManager struct {
	sessionFile string
	current     *Session
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	if err := sm.LoadSession(); err != nil {
		log.Printf("[DEBUG] No previous session found or failed to load: %v", err)
	}
	return sm
}

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

func (sm *SessionManager) LoadSession() error {
	data, err := os.ReadFile(sm.sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
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

func (sm *SessionManager) GetCurrentSession() *Session {
	return sm.current
}

func (sm *SessionManager) IsLoggedIn() bool {
	return sm.current != nil
}

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

func (sm *SessionManager) GetUserInfo() string {
	if sm.current == nil {
		return "Not logged in"
	}
	return fmt.Sprintf("Logged in as: %s (%s)", sm.current.Username, sm.current.Email)
}
