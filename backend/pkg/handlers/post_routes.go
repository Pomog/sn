package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/models"

	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid" // UUID package for generating and handling unique IDs
)

// Handler for inserting a new post
func insertPostHandler(ctx *socialnetwork.Context) {
	newPost := models.Post{}
	// Parse the incoming JSON body into the newPost struct
	if err := ctx.BodyParser(&newPost); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Error while creating new post",
		})
		return
	}

	// Retrieve the user ID from the request context
	userPostOwnerId := ctx.Values["userId"].(uuid.UUID)
	newPost.UserID = userPostOwnerId

	// Attempt to save the post in the database
	if err := newPost.Create(ctx.Db.Conn); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Error while creating new post",
		})
		return
	}

	// Send a successful response with the created post data
	ctx.JSON(map[string]interface{}{
		"status": http.StatusOK,
		"data":   newPost.ExploitForRendering(ctx.Db.Conn), // Render the post for frontend use
	})
}

// Handler for inserting a new comment
func insertCommentHandler(ctx *socialnetwork.Context) {
	newComment := models.Comment{}
	// Parse the incoming JSON body into the newComment struct
	if err := ctx.BodyParser(&newComment); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Error while creating new comment",
		})
		return
	}

	// Retrieve the user ID from the request context and associate it with the comment
	newComment.UserID = ctx.Values["userId"].(uuid.UUID)

	// Fetch the post to which the comment is being added (for validation or context)
	post := models.Post{}
	post.Get(ctx.Db.Conn, newComment.PostID)
	fmt.Println(post)

	// Attempt to save the comment in the database
	if err := newComment.Create(ctx.Db.Conn); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Error while creating new comment",
		})
		return
	}

	// Send a successful response with the created comment data
	ctx.JSON(map[string]interface{}{
		"status": http.StatusOK,
		"data":   newComment.PrepareForRendering(ctx.Db.Conn, string(post.Privacy), post.GroupID),
	})
}

// Handler for fetching the feed posts available to the user
func feedHandler(ctx *socialnetwork.Context) {
	feedPosts := models.Posts{}

	fmt.Println("func feedHandler(ctx *socialnetwork.Context) { *******************************************")
	fmt.Println(ctx)
	fmt.Printf("Type of userId: %T\n", ctx.Values["userId"])

	user := ctx.Values["userId"].(uuid.UUID)

	// Fetch posts accessible to the user based on their permissions
	if err := feedPosts.GetAvailablePostForUser(ctx.Db.Conn, user); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Error while getting posts",
		})
		return
	}

	// Send a successful response with the fetched posts
	ctx.JSON(map[string]interface{}{
		"status": http.StatusOK,
		"data":   feedPosts.ExploitForRendering(ctx.Db.Conn),
	})
}

// Handler for retrieving posts belonging to a specific group
func handleGetGroupPost(ctx *socialnetwork.Context) {
	// Extract group ID from query parameters
	id := ctx.Request.URL.Query().Get("id")
	post := models.Posts{}

	// Fetch posts associated with the specified group ID
	post.GetPostByGroupId(ctx.Db.Conn, id)

	// Send a successful response with the group posts
	ctx.JSON(map[string]interface{}{
		"status": http.StatusOK,
		"data":   post.ExploitForRendering(ctx.Db.Conn),
	})
}

// Route definitions with paths, HTTP methods, and middleware configurations
var getGroupsPostRoute = route{
	path:   "/post/groups",
	method: http.MethodGet,
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Ensure the user is authenticated
		handleGetGroupPost,      // Final handler for the route
	},
}

var insertPostRoute = route{
	path:   "/post/insert",
	method: http.MethodPost,
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Ensure the user is authenticated
		insertPostHandler,       // Final handler for the route
	},
}

var getFeedPostsRoute = route{
	path:   "/post/getFeed",
	method: http.MethodGet,
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Ensure the user is authenticated
		feedHandler,             // Final handler for the route
	},
}

var insertCommentRoot = route{
	path:   "/post/insertComment",
	method: http.MethodPost,
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Ensure the user is authenticated
		insertCommentHandler,    // Final handler for the route
	},
}

// Initialization function to register all defined routes
func init() {
	AllHandler[insertCommentRoot.path] = insertCommentRoot
	AllHandler[getFeedPostsRoute.path] = getFeedPostsRoute
	AllHandler[insertPostRoute.path] = insertPostRoute
	AllHandler[getGroupsPostRoute.path] = getGroupsPostRoute
}
