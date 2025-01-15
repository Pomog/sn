package router

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ApplyMiddleware chains multiple middleware functions around a handler.
// The last middleware function provided is the outermost.
func ApplyMiddleware(final http.Handler, funcs ...func(http.Handler) http.Handler) http.Handler {
	for _, fn := range funcs {
		final = fn(final) // Wrap the handler with each middleware.
	}
	return final // Return the wrapped handler.
}

// Recover catches panics in a handler and lets a custom handler handle the request instead.
// It logs the error if specified.
func Recover(next, recoverer http.Handler, logErr bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if logErr {
					log.Println(err) // Log the panic if requested.
					// debug.PrintStack() // Optional: Uncomment to print stack trace.
				}
				recoverer.ServeHTTP(w, r) // Serve the recovery handler.
			}
		}()

		next.ServeHTTP(w, r) // Proceed with the next handler.
	})
}

// Recover500 is a specific recovery handler that responds with a 500 Internal Server Error if the next handler panics.
func Recover500(next http.Handler) http.Handler {
	return Recover(
		next,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			code := http.StatusInternalServerError
			http.Error(w, fmt.Sprintf(`%v %v`, code, http.StatusText(code)), code) // Send a 500 error.
		}),
		true, // Enable error logging.
	)
}

// RedirectTrailingSlash ensures requests with a trailing slash are redirected to the same path without it.
func RedirectTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" && strings.HasSuffix(path, "/") {
			// Remove the trailing slash and redirect permanently.
			http.Redirect(w, r, path[:len(path)-1], http.StatusMovedPermanently)
			return
		}

		next.ServeHTTP(w, r) // Continue processing the request.
	})
}

// LogRequests logs the details of each incoming request (method, URI, and client address).
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := r.Header.Get("X-Forwarded-For")
		if client == "" {
			client = r.RemoteAddr // Use remote address if no forwarded address is provided.
		}
		log.Printf("%4s request for %v from %v\n", r.Method, r.URL.RequestURI(), client) // Log the request.
		next.ServeHTTP(w, r)                                                             // Process the request with the next handler.
	})
}
