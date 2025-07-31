package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// writeJSON is a helper for sending JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// readJSON is a helper for reading JSON requests
func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// apiError represents a standard error response
type apiError struct {
	Error string `json:"error"`
}

// permissionDenied sends a standard 403 Forbidden error
func permissionDenied(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, apiError{Error: "permission denied"})
}

// requireAuth is a middleware that validates JWT tokens and injects user info into request context
func requireAuth(s *APIServer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get JWT token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "authorization header required"})
			return
		}

		// Extract token from "Bearer <token>" format
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid authorization format"})
			return
		}
		tokenString := authHeader[len(bearerPrefix):]

		// Validate JWT token and get user info
		userInfo, err := s.validateJWT(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid token"})
			return
		}

		// Store user info in request context for the handler to use
		ctx := r.Context()
		ctx = SetUserInContext(ctx, userInfo)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// Context key for user info
type contextKey string

const userContextKey contextKey = "user"

// SetUserInContext stores user info in request context
func SetUserInContext(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext retrieves user info from request context
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(userContextKey).(*UserInfo)
	return user, ok
}
