package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isaacjstriker/devware/internal/auth"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	user, passwordHash, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		permissionDenied(w)
		return
	}

	if !auth.CheckPassword(req.Password, passwordHash) {
		permissionDenied(w)
		return
	}

	token, err := createJWT(user.ID, user.Username, s.config.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to create token"})
		return
	}

	resp := LoginResponse{
		Token:    token,
		Username: user.Username,
		UserID:   user.ID,
	}

	writeJSON(w, http.StatusOK, resp)
}

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

type UserInfo struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

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

	userID, ok := claims["userID"].(float64)
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

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}
