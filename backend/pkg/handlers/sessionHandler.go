package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"net/http"
)

// handleValidSession handles requests to validate whether a user's session is active and valid.
// This handler assumes that the request has passed through the required authentication middleware.
//
// Parameters:
// - ctx: The application's context object, which provides access to the HTTP request, response, and other shared resources.
func handleValidSession(ctx *socialnetwork.Context) {
	// Define the response data to be returned if the session is authenticated successfully.
	data := map[string]interface{}{
		"message": "Authenticated successfully", // Message indicating successful session validation.
	}

	// Respond with a JSON object containing the success message.
	ctx.JSON(data)
}

// checkSessionRoute defines the route for session validation.
//
// Properties:
// - path: The endpoint's URL path (e.g., "/checksession").
// - method: The HTTP method to be used (GET in this case).
// - middlewareAndHandler: A chain of middleware and the handler function to process the request.
var checkSessionRoute = route{
	path:   "/checksession", // Define the URL path for the session check.
	method: http.MethodGet,  // Use the GET method for this endpoint.
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Middleware to ensure the user is authenticated.
		/*
		   Additional middleware can be added here if required for further validations or processing.
		   For example:
		   - middleware.Logging for logging requests.
		   - middleware.RateLimiter for preventing excessive requests.
		*/
		handleValidSession, // Final handler to validate the session and return the response.
	},
}

// init registers the checkSessionRoute with the global AllHandler map.
// This ensures the route is available when the application initializes and starts handling requests.
func init() {
	AllHandler[checkSessionRoute.path] = checkSessionRoute // Add the route to the global route handler map.
}
