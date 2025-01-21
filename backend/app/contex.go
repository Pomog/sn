package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

// db is a wrapper around the database connection
type db struct {
	Conn *sql.DB
}

// Context represents the context of an HTTP request, including the response writer,
// request details, handlers, and shared data for the request lifecycle.
type Context struct {
	Db             *db                 // Database connection wrapper
	ResponseWriter http.ResponseWriter // HTTP response writer
	Request        *http.Request       // Incoming HTTP request
	handlers       []HandlerFunc       // Middleware or route handlers for the request
	index          int                 // Current handler index
	Values         map[any]any         // Shared values accessible during request handling
}

// BodyParser parses the request body into the specified struct or map.
// Returns an error if parsing fails.
func (c *Context) BodyParser(out interface{}) error {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %v", err)
		}
	}(c.Request.Body) // Ensure the body is closed after reading
	if err != nil {
		return err
	}

	// Unmarshal the JSON body into the provided output structure
	err = json.Unmarshal(body, &out)
	if err != nil {
		var e *json.SyntaxError
		if errors.As(err, &e) {
			log.Printf("Syntax error at byte offset %d", e.Offset)
		}
		return err
	}

	// Re-assign the request body for further consumption, if needed
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return nil
}

// JSON sends a JSON response to the client.
// Sets the Content-Type header to "application/json".
func (c *Context) JSON(data interface{}) error {
	c.ResponseWriter.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.ResponseWriter).Encode(data)
}

// Next executes the next handler in the chain, if available.
func (c *Context) Next() {
	if c.index < len(c.handlers) {
		handler := c.handlers[c.index]
		c.index++
		handler(c)
	}
}

// Render renders an HTML template with the provided data and sends it to the response writer.
// Returns an error if template parsing or execution fails.
func (c *Context) Render(path string, data interface{}) error {
	tp, err := template.ParseFiles(path)
	if err != nil {
		return err
	}
	return tp.Execute(c.ResponseWriter, data)
}

// Status sets the HTTP status code for the response.
func (c *Context) Status(code int) *Context {
	c.ResponseWriter.WriteHeader(code)
	return c
}

// WriteString writes a string response to the client.
func (c *Context) WriteString(s string) (int, error) {
	return c.ResponseWriter.Write([]byte(s))
}

// GetBearerToken extracts the Bearer token from the Authorization header.
// Returns an empty string if the token is not present or improperly formatted.
func (c *Context) GetBearerToken() string {
	var token string
	headerBearer := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}
	return token
}
