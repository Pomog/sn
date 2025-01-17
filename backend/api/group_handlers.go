package api

import (
	"database/sql"
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/router"
	"github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
	"time"
)

// GetAllGroups Retrieve all groups visible to the user (public and joined).
func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	userID := getPossibleUserID(r)

	groups, err := Database.Group.GetAll(userID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, groups)
}

// GetMyGroups Retrieve groups where the user is a member.
func GetMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := getPossibleUserID(r)

	groups, err := Database.Group.GetMyGroups(userID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, groups)
}

// GetGroupByID Fetch group details by its ID, including the user's membership status.
func GetGroupByID(w http.ResponseWriter, r *http.Request) {
	userID := getPossibleUserID(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	group, err := Database.Group.GetByID(groupID, userID)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	writeJSON(w, group)
}

// CreateGroup Create a new group with the current user as the owner.
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	group := models.Group{}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	group.OwnerID = session.UserID
	group.Type = "public" // Default all groups to public.

	id, err := Database.Group.Insert(group)
	panicUnlessError(err, sqlite3.ErrConstraintUnique)
	if err != nil {
		log.Println(err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	group.GroupID = id
	group.Created = time.Now()

	if err := Database.Group.Join(group.GroupID, group.OwnerID); err != nil {
		panic(err)
	}

	writeJSON(w, group)
}

// JoinGroup Join a group, or request access if the group requires it.
func JoinGroup(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	group, err := Database.Group.GetByID(groupID, session.UserID)
	panicIfErr(err)
	if group.IncludesMe {
		return
	}

	access, err := Database.Group.JoinCheck(groupID, session.UserID)
	panicIfErr(err)
	if access {
		panicIfErr(Database.Group.Join(groupID, session.UserID))
	} else {
		panicIfErr(Database.Group.Request(groupID, session.UserID))
		go func() {
			user, err := Database.User.GetByID(session.UserID)
			if err != nil {
				log.Println(err)
			}
			Notify.Request(group.Group, user)
		}()
	}
}

// LeaveGroup Leave a group the user is currently a member of.
func LeaveGroup(_ http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	if err := Database.Group.Leave(groupID, session.UserID); err != nil {
		panic(err)
	}
}

// GetGroupMembers Retrieve a list of all members in a specific group.
func GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
	members, err := Database.Group.GetMembers(groupID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, members)
}

// GroupInvite Send an invitation to another user to join a group.
func GroupInvite(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
	inviteeID, _ := strconv.ParseInt(router.GetSlug(r, 1), 10, 64)

	group, err := Database.Group.GetByID(groupID, session.UserID)
	if err != nil || !group.IncludesMe {
		writeStatusError(w, http.StatusForbidden)
		return
	}

	access, err := Database.Group.IncludesUser(groupID, inviteeID)
	if err != nil || access {
		return
	}

	requested, err := Database.Group.InviteCheck(groupID, inviteeID)
	panicIfErr(err)
	if requested {
		panicIfErr(Database.Group.Join(groupID, inviteeID))
	} else {
		panicIfErr(Database.Group.Invite(groupID, session.UserID, inviteeID))
		go Notify.Invite(group.Group, inviteeID)
	}
}

// TransferOwnership Transfer ownership of a group to another member.
func TransferOwnership(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
	newOwnerID, _ := strconv.ParseInt(router.GetSlug(r, 1), 10, 64)

	group, err := Database.Group.GetByID(groupID, session.UserID)
	panicUnlessError(err, sql.ErrNoRows)
	if err != nil || session.UserID != group.OwnerID {
		writeStatusError(w, http.StatusForbidden)
		return
	}

	newInGroup, err := Database.Group.IncludesUser(groupID, newOwnerID)
	if err != nil || !newInGroup {
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	panicIfErr(Database.Group.TransferOwnership(groupID, newOwnerID))
}

// GetPendingInvites Fetch pending invitations for a specific group.
func GetPendingInvites(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)
	members, err := Database.Group.GetPendingInvites(groupID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, members)
}
