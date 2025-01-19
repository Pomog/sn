package main

import (
	"github.com/Pomog/sn/backend/api"
	"github.com/Pomog/sn/backend/database"
	"github.com/Pomog/sn/backend/notify"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Ensure persistence directory exists for database storage.
	_ = os.Mkdir("./data_storage", os.ModePerm)

	// Initialize the database connection and dependency injection.
	db := database.NewDatabase("./data_storage/social_network.db")
	api.Database = db
	api.Notify = notify.NewNotifier(db)

	// Start a background process to clean expired sessions periodically.
	go cleanupExpiredSessions(db)

	// Initialize the router and set up API routes.
	rtr := router.New()

	// User-related endpoints
	rtr.Get("/user", api.EnsureAuth(api.GetUserBySession))
	rtr.Post("/user", api.EnsureAuth(api.UpdateUser))
	rtr.Get("/user/{id:[0-9]+}", api.OptionalAuth(api.GetUserByID))
	rtr.Get("/user/contacts", api.EnsureAuth(api.GetKnownUsers))
	rtr.Get("/user/{email}", api.GetUserByEmail)

	// Follower management endpoints
	rtr.Get("/user/{id:[0-9]+}/followers", api.UserFollowers)
	rtr.Get("/user/{id:[0-9]+}/following", api.UserFollowing)
	rtr.Post("/user/{id:[0-9]+}/follow", api.EnsureAuth(api.UserFollow))
	rtr.Post("/user/{id:[0-9]+}/unfollow", api.EnsureAuth(api.UserUnfollow))
	rtr.Post("/user/{id:[0-9]+}/accept", api.EnsureAuth(api.UserAcceptFollow))

	// Authentication endpoints
	rtr.Post("/register", api.Register)
	rtr.Post("/login", api.Login)
	rtr.Get("/logout", api.EnsureAuth(api.Logout))
	rtr.Get("/logout/all", api.EnsureAuth(api.LogoutAll))

	// Post-related endpoints
	rtr.Post("/post", api.EnsureAuth(api.CreatePost))
	rtr.Get("/post/{id:[0-9]+}", api.OptionalAuth(api.GetPostByID))
	rtr.Get("/posts", api.OptionalAuth(api.GetAllPosts))
	rtr.Get("/posts/groups", api.EnsureAuth(api.GetMyGroupPosts))
	rtr.Get("/posts/following", api.EnsureAuth(api.GetMyFollowingPosts))
	rtr.Get("/group/{id:[0-9]+}/posts", api.GroupAccessCheck(api.GetGroupPosts))
	rtr.Get("/user/{id:[0-9]+}/posts", api.OptionalAuth(api.GetUserPosts))

	// Comment-related endpoints
	rtr.Post("/post/{id:[0-9]+}/comment", api.EnsureAuth(api.CreateComment))
	rtr.Get("/post/{id:[0-9]+}/comments", api.OptionalAuth(api.GetCommentsByPost))

	// File upload/download endpoints
	rtr.Post("/file", api.FileUpload)
	rtr.Get("/file/{fileID:[a-zA-Z0-9-]+}", api.FileDownload)

	// Group management endpoints
	rtr.Post("/group", api.EnsureAuth(api.CreateGroup))
	rtr.Get("/groups", api.OptionalAuth(api.GetAllGroups))
	rtr.Get("/groups/my", api.EnsureAuth(api.GetMyGroups))
	rtr.Get("/group/{id:[0-9]+}", api.GroupAccessCheck(api.GetGroupByID))
	rtr.Post("/group/{id:[0-9]+}/join", api.EnsureAuth(api.JoinGroup))
	rtr.Post("/group/{id:[0-9]+}/leave", api.EnsureAuth(api.LeaveGroup))

	// Event-related endpoints
	rtr.Post("/event", api.EnsureAuth(api.CreateEvent))
	rtr.Post("/event/{id:[0-9]+}/attend", api.EnsureAuth(api.EventGoing))
	rtr.Get("/event/{id:[0-9]+}", api.EventAccessCheck(api.GetEvent))
	rtr.Get("/group/{id:[0-9]+}/events", api.GroupAccessCheck(api.GetGroupEvents))

	// Chat-related endpoints
	// Placeholder for WebSocket-based chat API integration
	// rtr.HandleFunc("/v1/chat", api.EnsureAuth(api.StartChatWebsocket))

	// Start the server and apply middleware.
	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router.ApplyMiddleware(
		rtr,
		api.ExtendSession,            // Middleware to extend session expiration
		router.RedirectTrailingSlash, // Middleware to handle trailing slashes in URLs
		router.LogRequests,           // Middleware to log HTTP requests
		router.Recover500,            // Middleware for centralized error recovery
	)))
}

// Background task to clean up expired sessions from the database.
func cleanupExpiredSessions(db *database.Database) {
	for {
		n, err := db.Session.CleanExpired()
		if err != nil {
			log.Printf("Error cleaning expired sessions: %v", err)
		} else {
			log.Printf("Removed %d expired sessions", n)
		}
		time.Sleep(24 * time.Hour)
	}
}
