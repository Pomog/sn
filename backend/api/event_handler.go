package api

import (
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/notify"
	"github.com/Pomog/sn/backend/router"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Notify *notify.Notifier

// CreateEvent handles the creation of a new event.
// Ensures the user has access to the associated group before creating the event.
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	event := &models.Event{}

	// Decode the event details from the request body.
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.Println("Error decoding event:", err)
		writeStatusError(w, http.StatusBadRequest)
		return
	}

	event.AuthorID = session.UserID

	// Validate the user's access to the group associated with the event.
	group, err := Database.Group.GetByID(event.GroupID, session.UserID)
	panicIfErr(err)
	if !group.IncludesMe {
		log.Printf("CreateEvent: User %v does not have access to group %v\n", session.UserID, event.GroupID)
		writeStatusError(w, http.StatusForbidden)
		return
	}

	// Insert the event into the database and set its metadata.
	id, err := Database.Event.Insert(*event)
	panicIfErr(err)

	event.EventID = id
	event.Created = time.Now()

	// Respond with the created event details.
	writeJSON(w, event)

	// Notify group members asynchronously.
	go func() {
		creator, err := Database.User.GetByID(session.UserID)
		if err != nil {
			log.Println("Error fetching creator details:", err)
		}

		members, err := Database.Group.GetMembers(group.GroupID)
		if err != nil {
			log.Println("Error fetching group members:", err)
		}

		Notify.EventCreated(group.Group, event, creator, members)
	}()
}

// EventGoing marks a user as attending an event.
// Ensures the user has access to the event before updating their status.
func EventGoing(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Validate if the user can join the event.
	access, err := Database.Event.CanJoin(eventID, session.UserID)
	panicIfErr(err)
	if !access {
		log.Printf("EventGoing: User %v is not part of event %v's group\n", session.UserID, eventID)
		writeStatusError(w, http.StatusForbidden)
		return
	}

	// Update the user's status to 'going' for the event.
	err = Database.Event.Going(eventID, session.UserID)
	panicIfErr(err)
}

// EventNotGoing marks a user as not attending an event.
// Ensures the user has access to the event before updating their status.
func EventNotGoing(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Validate if the user can join the event.
	access, err := Database.Event.CanJoin(eventID, session.UserID)
	panicIfErr(err)
	if !access {
		log.Printf("EventNotGoing: User %v is not part of event %v's group\n", session.UserID, eventID)
		writeStatusError(w, http.StatusForbidden)
		return
	}

	// Update the user's status to 'not going' for the event.
	err = Database.Event.NotGoing(eventID, session.UserID)
	panicIfErr(err)
}

// EventUnset removes the user's attendance status for an event.
func EventUnset(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Unset the user's status for the event.
	err := Database.Event.Unset(eventID, session.UserID)
	panicIfErr(err)
}

// GetEvent retrieves the details of a specific event.
// Ensures the user has access to view the event.
func GetEvent(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)
	eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Fetch event details from the database.
	event, err := Database.Event.GetByID(eventID, myID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, event)
}

// GetGroupEvents retrieves all events associated with a specific group.
// Ensures the user has access to the group before returning events.
func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	myID := getPossibleUserID(r)
	groupID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Fetch group events from the database.
	events, err := Database.Event.GetByGroup(groupID, myID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, events)
}

// GetEventMembers retrieves all members attending a specific event.
func GetEventMembers(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.ParseInt(router.GetSlug(r, 0), 10, 64)

	// Fetch event members from the database.
	members, err := Database.Event.GetMembers(eventID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, members)
}

// GetMyEvents retrieves all events associated with the authenticated user.
func GetMyEvents(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Fetch user's events from the database.
	events, err := Database.Event.GetByUser(session.UserID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, events)
}
