package api

import (
	"database/sql"
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreatePost handles the creation of a new post. It checks for proper access, validates privacy settings, and stores the post in the database.
func CreatePost(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Structure to hold the post data along with allowed users for manual privacy posts
	post := struct {
		models.Post
		AllowedUsers []int64 `json:"allowedUsers"`
	}{}

	// Decode the request body into the post structure
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Println(err)                           // Log the error if decoding fails
		writeStatusError(w, http.StatusBadRequest) // Respond with a bad request status
		return
	}

	// If the post belongs to a group, check if the user has access
	if post.GroupID != nil {
		post.Privacy = "public"

		// Check if the user is a member of the group
		access, err := Database.Group.IncludesUser(*post.GroupID, session.UserID)
		panicIfErr(err)
		if !access {
			writeStatusError(w, http.StatusForbidden) // Respond with forbidden if access is denied
			return
		}
	}

	// If the post has manual privacy, ensure that allowed users are provided
	if post.Privacy == "manual" {
		if post.AllowedUsers == nil {
			log.Println("Attempted to create a post with manual privacy but without allowed users")
			writeStatusError(w, http.StatusBadRequest) // Respond with bad request if no users are allowed
			return
		}

		// Validate the user IDs in the allowed users list
		for i, userID := range post.AllowedUsers {
			_, err := Database.User.GetByID(userID)
			if err != nil {
				log.Printf("Invalid user ID at allowedUsers[%v]: %v\n", i, userID)
				writeStatusError(w, http.StatusBadRequest) // Respond with bad request if any user ID is invalid
				return
			}
		}
	}

	// Validate the images associated with the post
	for _, img := range strings.Split(post.Images, ",") {
		if img == "" {
			continue
		}

		_, err = Database.File.Get(img)
		if err != nil {
			log.Printf("Could not find file with token %v\n", img)
			writeStatusError(w, http.StatusBadRequest) // Respond with bad request if any image is invalid
			return
		}
	}

	// Set the author of the post and insert it into the database
	post.AuthorID = session.UserID
	id, err := Database.Post.Insert(post.Post)
	if err != nil {
		panic(err)
	}

	// Insert the allowed users for manual privacy posts
	if post.Privacy == "manual" {
		for _, userID := range post.AllowedUsers {
			err = Database.Post.InsertAllowedUser(id, userID)
			panicIfErr(err)
		}
	}

	// Set additional post attributes and respond with the created post
	post.PostID = id
	post.Created = time.Now()

	writeJSON(w, post.Post)
}

// GetAllPosts retrieves all posts for the current user with pagination based on "beforeID".
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)

	beforeID, err := queryAtoi(r.URL.Query().Get("beforeID"))
	if err != nil {
		log.Println(err)                           // Log the error if parsing fails
		writeStatusError(w, http.StatusBadRequest) // Respond with bad request status
		return
	}

	posts, err := Database.Post.GetAll(myID, beforeID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, posts) // Respond with the list of posts
}

// GetPostByID retrieves a specific post by ID, checking if the current user has access.
func GetPostByID(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)
	slug := router.GetSlug(r, 0)
	postID, _ := strconv.ParseInt(slug, 10, 64)

	// Check if the user has access to the requested post
	allowed, err := Database.Post.HasAccess(myID, postID)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusNotFound) // Respond with not found if post doesn't exist
		return
	}
	if !allowed {
		writeStatusError(w, http.StatusForbidden) // Respond with forbidden if access is denied
		return
	}

	// Retrieve the post and respond with it
	post, err := Database.Post.GetByID(postID)
	panicIfErr(err)

	writeJSON(w, post)
}

// GetGroupPosts retrieves posts for a specific group.
func GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	beforeID, err := queryAtoi(r.URL.Query().Get("beforeID"))
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Respond with bad request status if parsing fails
		return
	}

	posts, err := Database.Post.GetByGroup(groupID, beforeID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, posts) // Respond with the list of group posts
}

// GetMyGroupPosts retrieves posts from the user's groups.
func GetMyGroupPosts(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	beforeID, err := queryAtoi(r.URL.Query().Get("beforeID"))
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Respond with bad request status if parsing fails
		return
	}

	posts, err := Database.Post.GetByMyGroups(session.UserID, beforeID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, posts) // Respond with the list of the user's group posts
}

// GetUserPosts retrieves posts made by a specific user.
func GetUserPosts(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)

	slug := router.GetSlug(r, 0)
	userID, _ := strconv.ParseInt(slug, 10, 64)

	beforeID, err := queryAtoi(r.URL.Query().Get("beforeID"))
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Respond with bad request status if parsing fails
		return
	}

	posts, err := Database.Post.GetByUser(myID, userID, beforeID)
	panicIfErr(err)

	writeJSON(w, posts) // Respond with the list of posts by the user
}

// GetMyFollowingPosts retrieves posts from users that the current user is following.
func GetMyFollowingPosts(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	beforeID, err := queryAtoi(r.URL.Query().Get("beforeID"))
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest) // Respond with bad request status if parsing fails
		return
	}

	posts, err := Database.Post.GetByFollowing(session.UserID, beforeID)
	panicIfErr(err)

	writeJSON(w, posts) // Respond with the list of posts from users the current user is following
}
