package session

import (
	socialnetwork "Social_Network/app"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notif is a sync.Map used to track session notification states for users.
// It keeps track of whether notifications are active for a particular user.
var Notif sync.Map

// Config defines the session configuration, including cookie settings.
type Config struct {
	CookieName string        // Name of the cookie to store session ID
	Value      string        // Value for the session cookie
	Path       string        // Path where the cookie is valid
	Domain     string        // Domain for the session cookie
	Expires    time.Time     // Expiration time for the session cookie
	RawExpires string        // Raw expiration value used during cookie reading
	MaxAge     int           // Max age for the session in seconds
	Secure     bool          // Whether the cookie is secure (HTTPS only)
	HttpOnly   bool          // If true, the cookie is inaccessible via JavaScript
	SameSite   http.SameSite // SameSite setting for cross-site cookie handling
	Raw        string        // Unparsed raw cookie string
	Unparsed   []string      // Additional unparsed cookie attributes
}

// starter represents the session starter that handles session-related logic.
type starter struct {
	session *session               // Reference to the session instance
	Ctx     *socialnetwork.Context // Context associated with the current session
}

// session holds session configuration, state, and the database connection.
type session struct {
	Config      *Config    // Session configuration
	database    *sql.DB    // Database connection for session persistence
	data        *sync.Map  // In-memory data storage for sessions
	mu          sync.Mutex // Mutex for synchronizing access to session data
	SessionName string     // Name of the session (e.g., "user_sessions")
}

// storage represents a session's cookie and ID, used for managing session state.
type storage struct {
	cookie *http.Cookie // Session cookie associated with the user
	id     uuid.UUID    // Unique identifier for the user session
}

// New initializes and returns a new session instance with default or provided configuration.
func New(c *Config) *session {
	if c == nil {
		c = new(Config) // If no config is provided, create a new empty config
	}
	if c.CookieName == "" {
		c.CookieName = "mycookie" // Default cookie name
	}
	if c.MaxAge == 0 {
		c.MaxAge = 31536000 // Default max age: 1 year
	}
	if c.Expires.IsZero() {
		c.Expires = time.Now().Add(time.Second * time.Duration(c.MaxAge)) // Default expiration
	}
	if c.SameSite == 0 {
		c.SameSite = http.SameSiteNoneMode // Default SameSite value
	}
	return &session{
		Config:      c,
		SessionName: c.CookieName,
		database:    nil,
		data:        &sync.Map{}, // Initialize the session data store
	}
}

// tmp is a background function that periodically removes expired sessions from memory and the database.
func (s *session) tmp() {
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Ticker for periodic cleanup
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					s.mu.Lock() // Lock to ensure safe concurrent access
					if s.database != nil {
						// Clean expired sessions from the database
						_, err := s.database.Exec(fmt.Sprintf(
							`DELETE FROM %s WHERE datetime(expiration_date) <= datetime('now')`,
							s.SessionName,
						))
						if err != nil {
							s.mu.Unlock()
							return
						}
					}

					// Clean expired sessions from in-memory storage
					s.data.Range(func(key, value interface{}) bool {
						val, ok := value.(map[string]interface{})
						if !ok {
							return true
						}
						userKey := val["key"].(uuid.UUID)
						storage := val["cookie"].(*storage)
						if time.Now().After(storage.cookie.Expires) {
							s.data.Delete(key)          // Remove expired session from memory
							Notif.Store(userKey, false) // Notify about session expiration
						}
						return true
					})
					s.mu.Unlock()
				case <-quit:
					ticker.Stop() // Stop the ticker when quitting
					return
				}
			}
		}()
		time.Sleep(30 * time.Minute) // Run cleanup every 30 minutes
		close(quit)
	}()
}

// UseDB associates a database with the session and creates the session table if it doesn't exist.
func (s *session) UseDB(db *sql.DB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.database = db
	_, err := s.database.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (id UUID PRIMARY KEY, user_id UUID NOT NULL, expiration_date DATETIME NOT NULL);`,
		s.SessionName,
	))
	if err != nil {
		log.Fatal(err) // Log error if database table creation fails
	}
}

// Start initializes a session starter that manages session-related operations.
func (s *session) Start(c *socialnetwork.Context) *starter {
	s.tmp() // Start periodic cleanup of expired sessions
	return &starter{session: s, Ctx: c}
}

// Set creates a new session for a user, storing it in the database and memory, and setting a cookie.
func (s *starter) Set(userID uuid.UUID) (string, error) {
	s.session.mu.Lock() // Lock session data for thread safety
	defer s.session.mu.Unlock()

	session := s.session
	db := session.database
	tmpdata := session.data

	// Generate a new session ID
	sessionID := uuid.New()

	// Check if a session already exists for the user and delete it if so
	if db != nil {
		var existingUserID uuid.UUID
		err := db.QueryRow(fmt.Sprintf(
			"SELECT id FROM %s WHERE user_id = $1", session.SessionName,
		), userID).Scan(&existingUserID)

		if err != sql.ErrNoRows {
			stmt, err := db.Prepare(fmt.Sprintf(
				"DELETE FROM %s WHERE id = $1", session.SessionName,
			))
			if err != nil {
				return "", err
			}
			_, err = stmt.Exec(existingUserID) // Remove the old session
			if err != nil {
				return "", err
			}
			tmpdata.Delete(existingUserID.String()) // Remove old session from memory
		}

		// Insert the new session for the user into the database
		stmt, err := db.Prepare(fmt.Sprintf(
			"INSERT INTO %s (id, user_id, expiration_date) VALUES ($1, $2, $3)",
			session.SessionName,
		))
		if err != nil {
			return "", err
		}
		_, err = stmt.Exec(sessionID, userID, time.Now().Add(time.Second*time.Duration(session.Config.MaxAge)))
		if err != nil {
			return "", err
		}
	}

	// Create a new session cookie for the user
	cookie := &http.Cookie{
		Name:     session.Config.CookieName,
		Value:    sessionID.String(),
		Secure:   session.Config.Secure,
		Expires:  time.Now().Add(time.Second * time.Duration(session.Config.MaxAge)),
		MaxAge:   session.Config.MaxAge,
		Path:     session.Config.Path,
		Domain:   session.Config.Domain,
		HttpOnly: session.Config.HttpOnly,
		SameSite: session.Config.SameSite,
	}

	// Set the cookie in the response header
	http.SetCookie(s.Ctx.ResponseWriter, cookie)

	// Store the new session in memory
	tmpdata.Store(sessionID.String(), map[string]interface{}{
		"key":    userID,
		"cookie": &storage{cookie: cookie, id: userID},
	})
	Notif.Store(userID, true) // Notify about the new session
	return sessionID.String(), nil
}

func (s *starter) Get(bearer string) (uuid.UUID, error) {
	s.session.mu.Lock()
	defer s.session.mu.Unlock()

	session := s.session
	c := session.Config
	db := session.database
	tmpdata := session.data

	// Retrieve the cookie
	cookie, err := s.Ctx.Request.Cookie(c.CookieName)
	if err != nil {
		// Fallback to the bearer token if the cookie is not found
		if bearer != "" {
			cookie = &http.Cookie{
				Name:  c.CookieName,
				Value: bearer,
			}
		} else {
			return uuid.Nil, fmt.Errorf("error retrieving cookie: %v", err)
		}
	}

	// Get the session ID from the cookie
	sessionID := cookie.Value

	// Check if the session exists in memory
	value, ok := tmpdata.Load(sessionID)
	if ok {
		val := value.(map[string]interface{})
		userKey := val["key"].(uuid.UUID)
		storage := val["cookie"].(*storage)

		// Check if the session is expired
		expirationDate := storage.cookie.Expires
		if time.Now().After(expirationDate) {
			tmpdata.Delete(sessionID)
			Notif.Store(userKey, false) // Notify that the session has expired
			return uuid.Nil, fmt.Errorf("session expired")
		} else {
			Notif.Store(userKey, true)
			return storage.id, nil
		}
	}

	// If the session is not in memory, check the database
	if db != nil {
		var userID uuid.UUID
		var expirationDate time.Time
		err = db.QueryRow(fmt.Sprintf("SELECT user_id, expiration_date FROM %s WHERE id = $1", session.SessionName), sessionID).Scan(&userID, &expirationDate)
		if err != nil {
			return uuid.Nil, fmt.Errorf("error retrieving session from database: %v", err)
		}

		// Check if the session is expired
		if time.Now().After(expirationDate) {
			Notif.Store(userID, false)
			return uuid.Nil, fmt.Errorf("session expired")
		}
		Notif.Store(userID, true)
		return userID, nil
	}

	// Return an error if no valid session is found
	return uuid.Nil, fmt.Errorf("session not found")
}

// Valid checks if the session is valid by verifying its expiration and existence in memory.
func (s *starter) Valid(sessionID string) bool {
	_, err := s.Get(sessionID)
	return err == nil // If no error, the session is valid
}

// Delete removes a session from memory and the database.
func (s *starter) Delete(sessionID string) error {
	s.session.mu.Lock()
	defer s.session.mu.Unlock()

	session := s.session
	db := session.database
	tmpdata := session.data

	// Retrieve session data and delete from memory
	sessionData, ok := tmpdata.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found")
	}
	data := sessionData.(map[string]interface{})
	cookie := data["cookie"].(*storage)

	// Delete session from the database
	if db != nil {
		stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id = $1", session.SessionName))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(cookie.id)
		if err != nil {
			return err
		}
	}

	// Delete session from memory
	tmpdata.Delete(sessionID)
	Notif.Store(cookie.id, false) // Notify session deletion
	return nil
}
