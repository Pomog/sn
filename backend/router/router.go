package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Router is a struct to define routes for handling requests.
// Routes can be added using methods like Get or Post.
// It picks the first matching route based on the registered order.
type Router struct {
	routes []route // List of all registered routes.
}

// route defines a single route with its method, pattern, and handler.
type route struct {
	method  string         // HTTP method (e.g., GET, POST).
	regex   *regexp.Regexp // Compiled regex for matching URL paths.
	handler http.Handler   // Function to handle the matched request.
}

// New initializes and returns an empty Router.
func New() Router {
	return Router{
		routes: []route{}, // Start with no routes.
	}
}

// Part 1: Adding routes

// Get adds a route for GET requests.
// The pattern is treated as a regex and allows capturing groups for dynamic paths.
func (router *Router) Get(pattern string, handler http.HandlerFunc) {
	router.addRoute("GET", pattern, handler)
}

// Post adds a route for POST requests.
// The pattern is treated as a regex and allows capturing groups for dynamic paths.
func (router *Router) Post(pattern string, handler http.HandlerFunc) {
	router.addRoute("POST", pattern, handler)
}

// addRoute registers a route with a specific HTTP method, pattern, and handler.
func (router *Router) addRoute(method string, pattern string, handler http.Handler) {
	rt := route{
		method:  method,
		regex:   regexp.MustCompile(fmt.Sprintf(`^%v$`, pattern)), // Compile the regex for the pattern.
		handler: handler,
	}

	router.routes = append(router.routes, rt) // Add the new route to the list.
}

// Part 2: Handling requests

// ServeHTTP processes incoming requests and finds a matching route.
// If a match is found, it calls the corresponding handler.
func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var allowed []string // Store allowed methods for the matched route, if any.

	// Check each route to find a match.
	for _, rt := range router.routes {
		match := rt.regex.FindStringSubmatch(r.URL.Path) // Check if the path matches the regex.

		if match == nil {
			// Skip to the next route if no match.
			continue
		}

		if rt.method != r.Method {
			// Save allowed methods if the method is different.
			allowed = append(allowed, rt.method)
			continue
		}

		if len(match) > 1 {
			// If capture groups exist, add them to the request context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, "routerSlugs", match[1:]) // Store the captured values in the context.
			r = r.WithContext(ctx)
		}

		// Call the handler for the matched route.
		rt.handler.ServeHTTP(w, r)
		return
	}

	// If no matching route was found.
	if len(allowed) > 0 {
		// Respond with 405 if a route matches but the method is not allowed.
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		code := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf(`%v %v`, code, http.StatusText(code)), code)
		return
	}

	// Respond with 404 if no route matches at all.
	http.NotFound(w, r)
}

// GetSlug gets a capture group value from the context by its index.
// Returns an empty string if the index is invalid or if no capture groups exist.
func GetSlug(r *http.Request, index int) string {
	slugs, ok := r.Context().Value("routerSlugs").([]string) // Get the saved capture groups from the context.
	if !ok || index >= len(slugs) {
		return "" // Return an empty string if invalid.
	}
	return slugs[index] // Return the value of the requested capture group.
}
