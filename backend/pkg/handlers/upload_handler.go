package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"net/http"
)

// handleUpload handles the logic for uploading an image.
// It retrieves the file URL from the context values (set by the ImageUploadMiddleware)
// and sends a JSON response with the uploaded file's URL.
func handleUpload(ctx *socialnetwork.Context) {
	ctx.JSON(map[string]interface{}{
		"imageurl": ctx.Values["file"], // Send the file URL back to the client.
	})
}

// UploadRoute defines the configuration for the "/upload" route.
// This route handles image uploads and ensures authentication before processing the request.
var UploadRoute = route{
	path:   "/upload",       // Endpoint path for uploading images.
	method: http.MethodPost, // HTTP method used for this endpoint (POST).
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,          // Middleware to enforce authentication.
		middleware.ImageUploadMiddleware, // Middleware to handle file upload logic.
		// Additional middleware can be added here if needed.
		handleUpload, // Final handler to process the upload request.
	},
}

// init initializes the upload route by registering it with the application's route map.
// This makes the route available for the application to handle incoming requests.
func init() {
	AllHandler[UploadRoute.path] = UploadRoute // Register the route in the global handler map.
}
