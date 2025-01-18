package api

import (
	"database/sql"
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"log"
	"net/http"
	"time"
)

// Login authenticates a user by verifying their credentials and returns a session token if valid.
func Login(w http.ResponseWriter, r *http.Request) {
	// Parse the provided credentials from the request body
	credentials := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		log.Println(err)                           // Log decoding error
		writeStatusError(w, http.StatusBadRequest) // Return bad request status
		return
	}

	// Retrieve user data from the database using the provided email
	user, err := Database.User.GetByEmail(credentials.Email)
	panicUnlessError(err, sql.ErrNoRows) // Handle potential no rows error
	if err != nil {
		writeStatusError(w, http.StatusUnauthorized) // Invalid credentials
		return
	}

	// Validate the user's password (Encryption TODO)
	if user.Password != credentials.Password {
		writeStatusError(w, http.StatusUnauthorized) // Invalid password
		return
	}

	// If credentials are valid, initiate a session for the user
	doLogin(w, user)
	writeJSON(w, user) // Return user data in response
}

// doLogin creates a session for the user and sets a session cookie.
func doLogin(w http.ResponseWriter, user *models.User) {
	token, err := Database.Session.Insert(user.UserID, sessionDuration)
	if err != nil {
		panic(err) // Handle unexpected errors
	}

	// Create and set a new session cookie
	cookie := newSessionCookie(token, time.Now().Add(sessionDuration))
	http.SetCookie(w, cookie)
}

// Logout terminates the user's session using the session token from the request.
func Logout(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Get the session token and remove the session
	token := session.Token

	success, err := Database.Session.Delete(token)
	if err != nil {
		panic(err) // Handle database deletion error
	}

	if !success {
		writeStatusError(w, http.StatusUnauthorized) // Unauthorized if session removal fails
		return
	}

	// Clear the session cookie
	cookie := newSessionCookie("deleted", time.Unix(0, 0))
	http.SetCookie(w, cookie)

	// Respond with no content status
	w.WriteHeader(http.StatusNoContent)
}

// LogoutAll logs the user out from all active sessions.
func LogoutAll(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Fetch the user based on the session user ID
	user, err := Database.User.GetByID(session.UserID)
	if err != nil {
		panic(err) // Handle database error
	}

	// Clear all sessions for the user
	err = Database.Session.ClearUser(user.UserID)
	if err != nil {
		panic(err) // Handle session clearing error
	}

	// Clear the session cookie
	cookie := newSessionCookie("deleted", time.Unix(0, 0))
	http.SetCookie(w, cookie)

	// Respond with no content status
	w.WriteHeader(http.StatusNoContent)
}

// Register creates a new user and logs them in immediately after registration.
func Register(w http.ResponseWriter, r *http.Request) {
	// POST /api/register

	// Define a custom struct to capture incoming user data (including password)
	incoming := models.UserIncoming{}

	// Parse the user data from the request body
	err := json.NewDecoder(r.Body).Decode(&incoming)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Return bad request status
		return
	}

	// Insert the new user into the database
	id, err := Database.User.Insert(incoming)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Handle insertion error
		return
	}

	// Retrieve the newly created user from the database
	user, err := Database.User.GetByID(id)
	if err != nil {
		panic(err) // Handle error when fetching user
	}

	// Log the user in after successful registration
	doLogin(w, user)
	writeJSON(w, user) // Return user data in the response
}
