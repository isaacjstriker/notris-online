package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON is a helper for sending JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// apiError represents a standard error response
type apiError struct {
	Error string `json:"error"`
}

// permissionDenied sends a standard 403 Forbidden error
func permissionDenied(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, apiError{Error: "permission denied"})
}
