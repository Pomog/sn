package cors

import (
	"Social_Network/app"
	"net/http"
	"strconv"
	"strings"
)

// Config defines the structure for CORS settings.
type Config struct {
	AllowedOrigins   []string // List of allowed origins (e.g., "*" for all).
	AllowedMethods   []string // List of allowed HTTP methods.
	AllowedHeaders   []string // List of allowed headers.
	AllowCredentials bool     // Whether to allow credentials.
	ExposedHeaders   []string // Headers exposed to the client.
	MaxAge           int      // Cache duration for preflight responses.
}

// DefaultConfig provides sensible default values for CORS.
func DefaultConfig() Config {
	return Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		AllowCredentials: false,
		ExposedHeaders:   []string{},
		MaxAge:           86400, // Default: 24 hours.
	}
}

// MergeConfig combines user-provided and default CORS settings.
func MergeConfig(userConfig Config) Config {
	defaultConfig := DefaultConfig()

	if len(userConfig.AllowedOrigins) > 0 {
		defaultConfig.AllowedOrigins = userConfig.AllowedOrigins
	}
	if len(userConfig.AllowedMethods) > 0 {
		defaultConfig.AllowedMethods = userConfig.AllowedMethods
	}
	if len(userConfig.AllowedHeaders) > 0 {
		defaultConfig.AllowedHeaders = userConfig.AllowedHeaders
	}
	if len(userConfig.ExposedHeaders) > 0 {
		defaultConfig.ExposedHeaders = userConfig.ExposedHeaders
	}
	if userConfig.MaxAge > 0 {
		defaultConfig.MaxAge = userConfig.MaxAge
	}
	defaultConfig.AllowCredentials = userConfig.AllowCredentials

	return defaultConfig
}

// New creates a CORS middleware handler based on the provided configuration.
func New(userConfig Config) app.HandlerFunc {
	config := MergeConfig(userConfig)

	return func(c *app.Context) {
		headers := c.ResponseWriter.Header()

		// Set CORS headers.
		headers.Set("Access-Control-Allow-Origin", strings.Join(config.AllowedOrigins, ","))
		headers.Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
		headers.Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(config.AllowCredentials))
		headers.Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ","))
		headers.Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

		// Handle preflight requests (OPTIONS).
		if c.Request.Method == http.MethodOptions {
			c.ResponseWriter.WriteHeader(http.StatusOK)
			return
		}

		// Proceed to the next middleware or handler.
		c.Next()
	}
}
