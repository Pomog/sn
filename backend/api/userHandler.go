package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"strconv"
)

// GetUserBySession Handles fetching user details based on the current session.
func GetUserBySession(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	user, err := Database.User.GetByID(session.UserID)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, user)
}

// GetUserByID Retrieves user details by their ID, considering session-based access rules.
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)
	slug := router.GetSlug(r, 0)

	id, _ := strconv.ParseInt(slug, 10, 64)

	user, err := Database.User.GetByIDPlusFollowInfo(id, myID)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Restricts access to private users if no follow relationship exists.
	if myID != id && (user.Private && !user.FollowInfo.MeToYou) {
		payload := struct {
			*models.UserLimited
			Access bool `json:"access"`
		}{
			UserLimited: user.Limited(),
			Access:      false,
		}

		writeJSON(w, payload)
		return
	}

	payload := struct {
		*models.User
		Access bool `json:"access"`
	}{
		User:   user,
		Access: true,
	}

	writeJSON(w, payload)
}

// GetUserByEmail Retrieves user details using their email address.
func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := router.GetSlug(r, 0)

	user, err := Database.User.GetByEmail(email)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil {
		writeStatusError(w, http.StatusNotFound)
		return
	}

	writeJSON(w, user)
}

// UpdateUser Updates user information based on incoming request data.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	user := models.UserIncoming{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	err = Database.User.Update(session.UserID, user)
	log.Printf("%T\n", errors.Unwrap(err))
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}
}

// UserFollow Allows the current user to follow another user.
func UserFollow(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	slug := router.GetSlug(r, 0)
	targetID, _ := strconv.ParseInt(slug, 10, 64)

	target, err := Database.User.GetByID(targetID)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	// Sends a follow request if the target user's profile is private.
	if target.Private {
		err = Database.User.RequestFollow(session.UserID, target.UserID)
		if err != nil {
			panic(err)
		}

		go func() {
			me, err := Database.User.GetByID(session.UserID)
			if err != nil {
				log.Println(err)
			}

			Notify.FollowRequest(me, targetID)
		}()
	} else {
		// Directly follows the user if their profile is public.
		err = Database.User.Follow(session.UserID, target.UserID)
		if err != nil {
			panic(err)
		}

		go func() {
			me, err := Database.User.GetByID(session.UserID)
			if err != nil {
				log.Println(err)
			}
			Notify.Follow(me, targetID)
		}()
	}
}

// UserAcceptFollow Accepts a follow request for the current user's private profile.
func UserAcceptFollow(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	slug := router.GetSlug(r, 0)
	targetID, _ := strconv.ParseInt(slug, 10, 64)

	err := Database.User.FollowAccept(session.UserID, targetID)
	if err != nil {
		panic(err)
	}

	go func() {
		me, err := Database.User.GetByID(session.UserID)
		if err != nil {
			log.Println(err)
		}

		Notify.FollowAccepted(me, targetID)
	}()
}

// UserUnfollow Removes a follow relationship between the current user and another user.
func UserUnfollow(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	slug := router.GetSlug(r, 0)
	targetID, _ := strconv.ParseInt(slug, 10, 64)

	err := Database.User.Unfollow(session.UserID, targetID)
	if err != nil {
		panic(err)
	}
}

// UserFollowers Fetches a list of followers for a specific user.
func UserFollowers(w http.ResponseWriter, r *http.Request) {
	slug := router.GetSlug(r, 0)
	targetID, _ := strconv.ParseInt(slug, 10, 64)

	users, err := Database.User.ListFollowers(targetID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, users)
}

// UserFollowing Fetches a list of users that a specific user is following.
func UserFollowing(w http.ResponseWriter, r *http.Request) {
	slug := router.GetSlug(r, 0)
	targetID, _ := strconv.ParseInt(slug, 10, 64)

	users, err := Database.User.ListFollowing(targetID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, users)
}

// GetKnownUsers Retrieves a list of users known to the current session's user.
func GetKnownUsers(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	users, err := Database.User.Known(session.UserID)
	panicIfErr(err)

	writeJSON(w, users)
}
