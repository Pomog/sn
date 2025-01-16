package api

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Pomog/sn/backend/database"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Database *database.Database

// ExtendSession refreshes the session expiration time if a valid session cookie exists.
func ExtendSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value

		success, err := Database.Session.SetExpires(token, sessionDuration)
		if err != nil {
			log.Println(fmt.Errorf("ExtendSession error: %w", err))
			next.ServeHTTP(w, r)
			return
		}

		if success {
			expires := time.Now().Add(sessionDuration)
			newCookie := newSessionCookie(token, expires)
			http.SetCookie(w, newCookie)
		}

		next.ServeHTTP(w, r)
	})
}

// IsAuth checks if the user is authenticated.
// Executes `yes` handler if authenticated; otherwise, executes `no` handler.
func IsAuth(yes http.HandlerFunc, no http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			no.ServeHTTP(w, r)
			return
		}

		token := cookie.Value
		session, err := Database.Session.Get(token)
		panicUnlessError(err, sql.ErrNoRows)
		if err != nil {
			no.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		yes.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuth allows the request to proceed regardless of authentication,
// but adds session data to the context if available.
func OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return IsAuth(next, next)
}

// EnsureAuth ensures that the user is authenticated.
// Sends a 401 Unauthorized response if the user is not authenticated.
func EnsureAuth(next http.HandlerFunc) http.HandlerFunc {
	return IsAuth(
		next,
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Unauthorized request to %v\n", r.URL.RequestURI())
			writeStatusError(w, http.StatusUnauthorized)
		},
	)
}

// GroupAccessCheck verifies if the user has access to a specific group.
// If access is denied, sends a 401 Unauthorized response.
func GroupAccessCheck(next http.HandlerFunc) http.HandlerFunc {
	return OptionalAuth(func(w http.ResponseWriter, r *http.Request) {
		groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
		userID := getPossibleUserID(r)

		access, err := Database.Group.HasAccess(groupID, userID)
		if err != nil {
			panic(err)
		}

		if !access {
			log.Printf("Access denied for group ID: %v\n", groupID)
			writeStatusError(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EventAccessCheck verifies if the user has access to a specific event.
// If access is denied, sends a 401 Unauthorized response.
func EventAccessCheck(next http.HandlerFunc) http.HandlerFunc {
	return OptionalAuth(func(w http.ResponseWriter, r *http.Request) {
		eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
		userID := getPossibleUserID(r)

		access, err := Database.Event.HasAccess(eventID, userID)
		if err != nil {
			panic(err)
		}

		if !access {
			log.Printf("Access denied for event ID: %v\n", eventID)
			writeStatusError(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
