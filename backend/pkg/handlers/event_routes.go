// Package handlers contains the handler functions for various routes.
package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/models"

	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// createEventHandler handles the creation of a new event in a group.
func createEventHandler(ctx *socialnetwork.Context) {
	newEvent := models.Event{}

	// Parse the request body into the newEvent object
	if err := ctx.BodyParser(&newEvent); err != nil {
		// Handle error if body parsing fails
		ctx.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Set the creator and group IDs from the context values
	newEvent.CreatorID = ctx.Values["userId"].(uuid.UUID)
	newEvent.GroupID = ctx.Values["group_id"].(uuid.UUID)

	// Create the event in the database
	if err := newEvent.Create(ctx.Db.Conn); err != nil {
		// Handle error if event creation fails
		ctx.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Fetch group members to send notifications
	members := new(models.GroupMembers)
	if err := members.Get(ctx.Db.Conn, newEvent.GroupID, models.MemberStatusAccepted); err != nil {
		// Handle error if fetching members fails
		ctx.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Get group details
	group := ctx.Values["group"].(*models.Group)

	// Notify all group members about the new event
	for _, member := range *members {
		log.Println(member.ID)
		notif := &models.Notification{
			UserID:    newEvent.CreatorID,
			ConcernID: member.MemberID,
			MemberId:  member.ID,
			GroupId:   newEvent.GroupID,
			Message:   fmt.Sprintf("add new event in the group %s", group.Title),
			Type:      models.TypeNewEvent,
		}

		// Create notification for the member
		if err := notif.Create(ctx.Db.Conn); err != nil {
			log.Println(err)
			continue // Continue on error, but do not stop execution
		}
		log.Println(notif.ID)
	}

	// Return success response with the created event
	ctx.Status(http.StatusCreated).JSON(map[string]interface{}{
		"message": "Event created successfully",
		"data":    newEvent,
	})
}

// Define the route for creating an event
var createEventRoute = route{
	path:   "/create-event", // Path for creating event
	method: http.MethodPost, // HTTP method (POST)
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,    // Authentication required
		middleware.IsGroupExist,    // Check if the group exists
		middleware.HaveGroupAccess, // Check if the user has access to the group
		createEventHandler,         // Event handler
	},
}

// getAllEventByGroup handles fetching all events in a group.
func getAllEventByGroup(ctx *socialnetwork.Context) {
	events := models.Events{}
	groupId := ctx.Values["group_id"].(uuid.UUID)
	// Check query parameters for additional filters
	isParticipantNeeded := ctx.Request.URL.Query().Get("isParticipantNeeded") == "true"
	isUserNeeded := ctx.Request.URL.Query().Get("isUserNeeded") == "true"

	// Fetch the events from the database
	err := events.GetGroupEvents(ctx.Db.Conn, groupId, isParticipantNeeded, isUserNeeded)
	if err != nil {
		// Handle error if fetching events fails
		ctx.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Return success response with the events
	ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "All events",
		"data":    events,
	})
}

// Define the route for getting all events in a group
var getAllEventRoute = route{
	path:   "/get-all-event-group", // Path for getting events
	method: http.MethodGet,         // HTTP method (GET)
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,    // Authentication required
		middleware.IsGroupExist,    // Check if the group exists
		middleware.HaveGroupAccess, // Check if the user has access to the group
		getAllEventByGroup,         // Event handler
	},
}

// respondEventHandler handles a user's response to an event (going or not going).
func respondEventHandler(ctx *socialnetwork.Context) {
	event := ctx.Values["event"].(*models.Event)
	member := ctx.Values["member"].(*models.GroupMember)

	participant := models.EventParticipant{}
	_participant := models.EventParticipant{}

	// Parse the participant response from the request body
	if err := ctx.BodyParser(&_participant); err != nil {
		// Handle error if body parsing fails
		ctx.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Validate the participant response (should be either 'going' or 'not_going')
	if _participant.Response != "not_going" && _participant.Response != "going" {
		ctx.Status(http.StatusBadRequest).JSON("Invalid response")
		return
	}

	// Fetch existing participant data or create a new participant record
	err := participant.GetParticipant(ctx.Db.Conn, event.ID, member.ID, member.MemberID, false)
	participant.Response = _participant.Response
	if err != nil {
		// Create new participant if not found
		err := participant.CreateParticipant(ctx.Db.Conn, event.ID, member.ID)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	} else {
		// Update participant response if already exists
		err := participant.UpdateParticipant(ctx.Db.Conn)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	// Return success response with the updated participant data
	ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "Response updated",
		"data":    participant,
	})
}

// Define the route for responding to an event
var respondEventRoute = route{
	path:   "/response-event", // Path for responding to an event
	method: http.MethodPost,   // HTTP method (POST)
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,    // Authentication required
		middleware.IsGroupExist,    // Check if the group exists
		middleware.HaveGroupAccess, // Check if the user has access to the group
		middleware.IsEventExist,    // Check if the event exists
		respondEventHandler,        // Event handler
	},
}

// Initialize routes for the application
func init() {
	AllHandler[createEventRoute.path] = createEventRoute
	AllHandler[getAllEventRoute.path] = getAllEventRoute
	AllHandler[respondEventRoute.path] = respondEventRoute
}
