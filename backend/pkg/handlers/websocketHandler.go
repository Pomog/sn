package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/models"

	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ConnWrapper wraps a WebSocket connection, adding metadata such as its "Closed" state.
type ConnWrapper struct {
	Conn   *websocket.Conn // The WebSocket connection object.
	Closed bool            // Indicates whether the connection is closed.
}

// ConnMap is a thread-safe map for storing active WebSocket connections.
type ConnMap struct {
	sync.Map // Embedded sync.Map for concurrent access to connections.
}

var (
	// upgrader configures WebSocket connections, including buffer sizes.
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conns = ConnMap{} // Global map to store all active WebSocket connections.
	once  sync.Once   // Ensures initialization logic is executed only once.
)

// sendErrorAndClose sends an error message to the client and closes the WebSocket connection.
// Parameters:
// - conn: The WebSocket connection to close.
// - status: HTTP status code for the error.
// - message: Error message to send.
// - id: The UUID of the connection to remove from the global map.
func sendErrorAndClose(conn *websocket.Conn, status int, message string, id uuid.UUID) {
	err := conn.WriteJSON(map[string]interface{}{
		"status":  status,
		"message": message,
	})
	if err != nil {
		log.Printf("Error writing to WebSocket: %v", err)
	}
	conn.Close()
	conns.Delete(id)
}

// handleSocket handles incoming WebSocket connections and manages real-time messaging.
func handleSocket(ctx *socialnetwork.Context) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(ctx.ResponseWriter, ctx.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Assign a unique ID to the connection and store it in the global map.
	id := uuid.New()
	conns.Store(id, &ConnWrapper{Conn: conn, Closed: false})

	// Start a goroutine for broadcasting data to all active clients.
	once.Do(func() {
		go func() {
			log.Println("Starting WebSocket broadcast goroutine")
			for {
				// Wait for new data to be available in the models.Data channel.
				value := <-models.Data
				key, ok := value["key"].(string)
				data, okData := value["data"].(map[string]interface{})

				if ok && okData {
					// Iterate over all active connections.
					conns.Range(func(k, v interface{}) bool {
						connID, validID := k.(uuid.UUID)
						connWrapper, validWrapper := v.(*ConnWrapper)

						if validID && validWrapper && !connWrapper.Closed {
							// Set a Pong handler to verify the connection is still alive.
							connWrapper.Conn.SetPongHandler(func(string) error {
								connWrapper.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
								return nil
							})

							// Send a Ping message to keep the connection active.
							if err := connWrapper.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
								connWrapper.Closed = true
								conns.Delete(connID)
								return true
							}

							// Send the actual data to the client.
							if err := connWrapper.Conn.WriteJSON(map[string]interface{}{
								"data": data,
								"type": key,
							}); err != nil {
								connWrapper.Closed = true
								conns.Delete(connID)
							}
						}
						return true
					})
				}
			}
		}()
	})

	// Handle incoming messages from the client.
	for {
		var incomingData map[string]interface{}
		if err := conn.ReadJSON(&incomingData); err != nil {
			sendErrorAndClose(conn, http.StatusBadRequest, "Invalid message format", id)
			return
		}

		// Handle specific message types, e.g., private or group messages.
		switch incomingData["type"] {
		case "private_message":
			handlePrivateMessage(ctx, conn, incomingData, id)
		case "group_message":
			// Add implementation for handling group messages here.
		default:
			sendErrorAndClose(conn, http.StatusBadRequest, "Unknown message type", id)
			return
		}
	}
}

// handlePrivateMessage processes a private message received over the WebSocket connection.
// Parameters:
// - ctx: The application's context object.
// - conn: The WebSocket connection sending the message.
// - incomingData: The data payload received from the client.
// - id: The unique identifier for the WebSocket connection.
func handlePrivateMessage(ctx *socialnetwork.Context, conn *websocket.Conn, incomingData map[string]interface{}, id uuid.UUID) {
	msg, ok := incomingData["message"].(map[string]interface{})
	if !ok {
		sendErrorAndClose(conn, http.StatusBadRequest, "Invalid message format", id)
		return
	}

	privateMessage := models.PrivateMessage{
		Content:    msg["content"].(string),
		SenderID:   uuid.MustParse(msg["sender_id"].(string)),
		ReceiverID: uuid.MustParse(msg["receiver_id"].(string)),
	}

	// Validate the private message fields.
	if privateMessage.Content == "" || privateMessage.SenderID == uuid.Nil || privateMessage.ReceiverID == uuid.Nil {
		sendErrorAndClose(conn, http.StatusBadRequest, "Incomplete message data", id)
		return
	}

	user := models.User{}
	// Verify the existence of sender and receiver users.
	if user.Get(ctx.Db.Conn, privateMessage.ReceiverID) != nil || user.Get(ctx.Db.Conn, privateMessage.SenderID) != nil {
		sendErrorAndClose(conn, http.StatusNotFound, "Sender not found", id)
		return
	}

	// Save the private message to the database.
	if err := privateMessage.Create(ctx.Db.Conn); err != nil {
		sendErrorAndClose(conn, http.StatusInternalServerError, "Failed to save message", id)
		return
	}

	// Create a notification for the message sender.
	notification := models.Notification{
		UserID:    privateMessage.SenderID,
		ConcernID: privateMessage.ReceiverID,
		Type:      models.TypeNewMessage,
		Message:   privateMessage.Content,
	}
	if err := notification.Create(ctx.Db.Conn); err != nil {
		sendErrorAndClose(conn, http.StatusInternalServerError, "Failed to create notification", id)
	}
}

// handleSocketRoute defines the WebSocket endpoint for handling client connections.
var handleSocketRoute = route{
	path:   "/socket", // Endpoint URL path for WebSocket connections.
	method: http.MethodGet,
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AllowedServer, // Middleware to validate the request origin/server.
		handleSocket,             // WebSocket handler for real-time communication.
	},
}

// init registers the WebSocket route with the global AllHandler map.
func init() {
	AllHandler[handleSocketRoute.path] = handleSocketRoute
}
