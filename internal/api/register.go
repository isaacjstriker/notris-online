package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/isaacjstriker/devware/internal/auth"
)

type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid request body"})
		return
	}

	if err := auth.ValidateUsername(req.Username); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if err := auth.ValidateEmail(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to hash password"})
		return
	}

	user, err := s.db.CreateUser(req.Username, req.Email, hashedPassword)
	if err != nil {
		log.Printf("Error creating user: %v", err)

		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			writeJSON(w, http.StatusConflict, apiError{Error: "username or email already exists"})
		} else {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to create user"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}
