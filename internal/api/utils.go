package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type apiError struct {
	Error string `json:"error"`
}

func permissionDenied(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, apiError{Error: "permission denied"})
}

func requireAuth(s *APIServer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "authorization header required"})
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid authorization format"})
			return
		}
		tokenString := authHeader[len(bearerPrefix):]

		userInfo, err := s.validateJWT(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid token"})
			return
		}

		ctx := r.Context()
		ctx = SetUserInContext(ctx, userInfo)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

type contextKey string

const userContextKey contextKey = "user"

func SetUserInContext(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(userContextKey).(*UserInfo)
	return user, ok
}
