package middleware

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/config"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// AuthRequired checks if the user is authenticated
func AuthRequired(ctx *socialnetwork.Context) {
	var token string
	headerBearer := ctx.Request.Header.Get("Authorization")
	// Extract the token if the header contains "Bearer "
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}

	// Check if the session is valid for the provided token
	if !config.Sess.Start(ctx).Valid(token) {
		// Respond with an error if the user is not authenticated
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authenticated.",
		})
		return
	}

	// Retrieve the user ID from the session
	userId, err := config.Sess.Start(ctx).Get(token)
	fmt.Println("// Retrieve the user ID from the session")
	fmt.Println(userId)

	if err != nil {
		// Respond with an error if the user is not authenticated
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authenticated.",
		})
		return
	}

	// Store user ID and token in context for further use
	ctx.Values["userId"] = userId
	ctx.Values["token"] = token
	// Proceed to the next middleware
	ctx.Next()
}

// NoAuthRequired checks if the user is not authenticated
func NoAuthRequired(ctx *socialnetwork.Context) {
	var token string
	headerBearer := ctx.Request.Header.Get("Authorization")
	// Extract the token if the header contains "Bearer "
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}

	// Check if the session is already valid for the provided token
	if config.Sess.Start(ctx).Valid(token) {
		// Respond with an error if the user is already authenticated
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are already authenticated.",
		})
		return
	}
	// Proceed to the next middleware
	ctx.Next()
}

// AllowedServer checks if the request is coming from an authorized server
func AllowedServer(ctx *socialnetwork.Context) {
	key := ctx.Request.URL.Query().Get("key")
	// Check if the provided server key matches the expected one
	if key != os.Getenv("SERVER_KEY") {
		// Respond with an error if the server is not authorized
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not allowed to access this server.",
		})
		return
	}
	// Proceed to the next middleware
	ctx.Next()
}
