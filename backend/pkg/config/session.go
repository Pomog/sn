package config

import (
	"Social_Network/app/session"
)

// DefaultSessionConfig defines the default configuration for session management.
func DefaultSessionConfig() session.Config {
	return session.Config{
		CookieName: "sessions",
	}
}

// Initialize session with the default configuration.
var conf = DefaultSessionConfig() // Create a variable for the configuration
var Sess = session.New(&conf)     // Pass the address of the configuration variable
