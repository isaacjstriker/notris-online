package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isaacjstriker/devware/internal/auth"
)

// LoginRequest defines the shape of the login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the shape of the successful login response
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// handleLogin handles user login and JWT generation
func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	// Fetch user from database
	user, passwordHash, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		permissionDenied(w)
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, passwordHash) {
		permissionDenied(w)
		return
	}

	// Create JWT
	token, err := createJWT(user.ID, user.Username, s.config.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to create token"})
		return
	}

	resp := LoginResponse{
		Token:    token,
		Username: user.Username,
	}

	writeJSON(w, http.StatusOK, resp)
}

// createJWT generates a new JWT for a given user
func createJWT(userID int, username, secret string) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 1 week
		"issuedAt":  jwt.NewNumericDate(time.Now()),
		"userID":    userID,
		"username":  username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// UserInfo represents validated user information from JWT
type UserInfo struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// validateJWT validates a JWT token and returns user information
func (s *APIServer) validateJWT(tokenString string) (*UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenNotValidYet
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	userID, ok := claims["userID"].(float64) // JSON numbers are float64
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return &UserInfo{
		UserID:   int(userID),
		Username: username,
	}, nil
}

// handleLogout handles user logout
func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	// For JWT, logout is handled client-side by deleting the token
	// We could maintain a blacklist in the future if needed
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}
