package handlers

import (
	socialnetwork "Social_Network/app"
	"net/http"
)

// HandlerConstructor is a type alias for a function that takes a path and a variadic number of HandlerFuncs.
// It's used to define the constructor for creating new routes with associated middleware and handlers.
type HandlerConstructor func(path string, middlewareAndHandler ...socialnetwork.HandlerFunc)

// route represents a route with its associated path, constructor, and middleware/handler functions.
type route struct {
	path, method         string
	middlewareAndHandler []socialnetwork.HandlerFunc
}

// AllHandler is a map that defines all the routes for the application.
// Each route includes the path, the constructor for creating the route, and the middleware/handler functions to be executed.
var AllHandler = map[string]route{}

// HandleAll is a function that iterates over the AllHandler map and applies each Handler's constructor to register the routes.
// This function should be called during the initialization phase of the application to set up all the routes.
var HandleAll = func(app *socialnetwork.App) {
	// Use a map of HTTP methods to corresponding route constructors
	var mapConstructors = map[string]HandlerConstructor{
		http.MethodGet:    app.GET,
		http.MethodDelete: app.DELETE,
		http.MethodPost:   app.POST,
		http.MethodPut:    app.PUT,
	}

	// Iterate over the AllHandler map and apply the constructors
	for _, v := range AllHandler {
		mapConstructors[v.method](v.path, v.middlewareAndHandler...) // Apply the method-specific constructor
	}
}
