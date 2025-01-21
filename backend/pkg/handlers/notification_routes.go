// Package handlers contains the handler functions for various routes related to notifications.
package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/models"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// handlerNotifications retrieves and returns a list of notifications for the authenticated user.
func handlerNotifications(ctx *socialnetwork.Context) {
	// Get the user ID from the context
	userId := ctx.Values["userId"].(uuid.UUID)

	// Fetch notifications for the user
	notifications := new(models.Notifications)
	if err := notifications.GetByUser(ctx.Db.Conn, userId); err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": err,
		})
		return
	}

	// Prepare the response with the relevant notification data
	AllNotification := []map[string]interface{}{}
	for _, notification := range *notifications {

		// Retrieve the user details for the notification
		user := new(models.User)
		user.Get(ctx.Db.Conn, notification.UserID)
		user.Password = "" // Do not expose user password

		// Append each notification to the response list
		AllNotification = append(AllNotification, map[string]interface{}{
			"id":         notification.ID,
			"type":       notification.Type,
			"concernID":  notification.ConcernID,
			"user":       user,
			"message":    notification.Message,
			"created_at": notification.CreatedAt,
			"group_id":   notification.GroupId,
			"member_id":  notification.MemberId,
			"is_invite":  notification.Is_invite,
		})
	}

	// Return the notifications to the client
	ctx.JSON(AllNotification)
}

// handlerclearnotifications handles the clearing of notifications, either single or all.
func handlerclearnotifications(ctx *socialnetwork.Context) {
	// Get the user ID from the context
	userId := ctx.Values["userId"].(uuid.UUID)

	// Define the structure for incoming requests to clear notifications
	type request struct {
		Type   string `json:"type"`   // Type of request ("clear" or "clear_all")
		Id     string `json:"id"`     // Notification ID if clearing a single notification
		Action string `json:"action"` // Action for clearing (e.g., "new_message", "follow")
	}

	// Parse the incoming request
	req := new(request)
	if err := ctx.BodyParser(req); err != nil {
		ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid request",
		})
		return
	}

	// Process clearing of a single notification
	if req.Type == "clear" {
		notification := new(models.Notification)
		if err := notification.Get(ctx.Db.Conn, uuid.MustParse(req.Id)); err != nil {
			ctx.Status(http.StatusNotFound).JSON(map[string]interface{}{
				"error": "Notification not found",
			})
			return
		}

		// Ensure the user is authorized to clear the notification
		if notification.ConcernID != userId {
			ctx.Status(http.StatusForbidden).JSON(map[string]interface{}{
				"error": "You are not authorized to clear this notification",
			})
			return
		}

		// Delete the notification
		if err := notification.Delete(ctx.Db.Conn); err != nil {
			ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
				"error": "Something went wrong",
			})
			return
		}

		// Respond with success message
		ctx.JSON(map[string]interface{}{
			"message": "Notification cleared",
		})
		return
	} else if req.Type == "clear_all" {
		// Clear all notifications for the user
		notifications := new(models.Notifications)
		if err := notifications.GetByUser(ctx.Db.Conn, userId); err != nil {
			ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
				"error": "Something went wrong",
			})
			return
		}

		// Loop through each notification and delete as needed based on the action type
		for _, notification := range *notifications {
			if notification.Type == models.TypeNewMessage && req.Action != "new_message" {
				continue
			} else if strings.HasPrefix(string(notification.Type), "follow") && req.Action != "follow" {
				continue
			}

			// Delete the notification
			if err := notification.Delete(ctx.Db.Conn); err != nil {
				ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
					"error": "Something went wrong",
				})
				return
			}
		}

		// Respond with success message
		ctx.JSON(map[string]interface{}{
			"message": "Notifications cleared",
		})
		return
	} else {
		// Invalid request type
		ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid request",
		})
		return
	}
}

// Route for fetching notifications with authentication middleware
var notificationsRoute = route{
	path:   "/getnotifications", // Endpoint to fetch notifications
	method: http.MethodGet,      // HTTP method
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired, // Ensure user is authenticated
		handlerNotifications,    // The handler function to call
	},
}

// Route for clearing notifications with authentication middleware
var clearnotificationsRoute = route{
	path:   "/clearnotifications", // Endpoint to clear notifications
	method: http.MethodPost,       // HTTP method
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,   // Ensure user is authenticated
		handlerclearnotifications, // The handler function to call
	},
}

// Initialization of routes to be registered
func init() {
	// Register the notification-related routes
	AllHandler[notificationsRoute.path] = notificationsRoute
	AllHandler[clearnotificationsRoute.path] = clearnotificationsRoute
}
