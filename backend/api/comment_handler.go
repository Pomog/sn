package api

import (
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"strconv"
	"time"
)

// CreateComment handles the creation of a new comment on a post.
// Validates access to the post before saving the comment.
func CreateComment(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	postID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	comment := models.Comment{}

	// Decode the comment data from the request body.
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		log.Println("Error decoding comment:", err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	// Verify if the user has access to the post.
	access, err := Database.Post.HasAccess(session.UserID, postID)
	panicIfErr(err)
	if !access {
		log.Println("Access denied to the post")
		writeStatusError(w, http.StatusForbidden)
		return
	}

	// Populate comment details and save it to the database.
	comment.PostID = postID
	comment.AuthorID = session.UserID

	id, err := Database.Comment.Insert(comment)
	panicIfErr(err)

	// Complete comment metadata and send the response.
	comment.CommentID = id
	comment.Created = time.Now()

	writeJSON(w, comment)
}

// GetCommentsByPost retrieves all comments associated with a specific post.
// Validates access to the post before returning the comments.
func GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)
	postID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Verify if the user has access to the post.
	access, err := Database.Post.HasAccess(myID, postID)
	panicIfErr(err)
	if !access {
		log.Println("Access denied to the post")
		writeStatusError(w, http.StatusForbidden)
		return
	}

	// Fetch comments from the database and send them as a response.
	comments, err := Database.Comment.GetByPost(postID)
	panicIfErr(err)

	writeJSON(w, comments)
}
