package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"net/http"
	"strconv"
	"time"
)

const sessionDuration = 24 * time.Hour

// writeJSON sends a JSON response to the client. Triggers a panic if encoding fails.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		panic(err) // Stop execution if JSON encoding fails.
	}
}

// panicUnlessError checks the given error and triggers a panic if it is not nil,
// unless the error matches one of the allowed exceptions.
func panicUnlessError(err error, unless ...error) {
	if err == nil || errorInList(err, unless) {
		return
	}
	panic(err) // Panics if the error doesn't match allowed exceptions.
}

// errorInList checks if the provided error matches any error in the list.
func errorInList(err error, checks []error) bool {
	for _, check := range checks {
		if errors.Is(err, check) {
			return true
		}
	}
	return false
}

// writeStatusError sends an HTTP error response with the given status code.
func writeStatusError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf(`%v %v`, code, http.StatusText(code)), code)
}

// newSessionCookie creates a new HTTP cookie for managing session tokens.
func newSessionCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true, // Cookie is accessible only to the server.
	}
}

// getSession retrieves the session object from the request context.
// Panics if the session is unavailable or the handler is not authenticated.
func getSession(r *http.Request) *models.Session {
	session, ok := r.Context().Value("session").(*models.Session)
	if !ok {
		panic("Session is not available. Ensure the handler is authenticated using api.IsAuth and api.EnsureAuth.")
	}

	return session
}

// getPossibleUserID extracts the user ID from the session if available.
// Returns -1 if the session is missing or invalid.
func getPossibleUserID(r *http.Request) int64 {
	var userID int64 = -1
	session, ok := r.Context().Value("session").(*models.Session)
	if ok {
		userID = session.UserID
	}
	return userID
}

// panicIfErr triggers a panic if the provided error is not nil.
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// queryAtoi converts a query string to an integer.
// Returns -1 if the string is empty or conversion fails, along with an error.
func queryAtoi(s string) (int64, error) {
	if s == "" {
		return -1, nil
	}

	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid integer: %v", s)
	}

	return id, nil
}
