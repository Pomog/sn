// Package notify implements the notification service for handling and delivering
// real-time messages and updates to users. It includes the Notifier struct and
// its associated methods for processing notifications, formatting them, and
// delivering them to specified targets.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Pomog/sn/backend/database"
	"github.com/Pomog/sn/backend/models"
	"github.com/gorilla/websocket"
	"html"
	"log"
	"net/http"
	"os"
	"time"
)

// Retrieves the frontend host address from environment variables, defaulting to "localhost" if not set.
var frontendHost = getFrontendHost()

// Notification defines the interface for creating notifications.
// Includes methods for specifying target recipients, the message body, and associated links.
type Notification interface {
	Targets() []int64
	Message() string
	Links() []Link
}

// Link represents an actionable button in the notification,
// including its display name, URL, and HTTP method.
type Link struct {
	name   string
	url    string
	method string
}

// String generates an HTML representation of the Link for rendering in notifications.
func (l Link) String() string {
	return fmt.Sprintf(
		"\n<button type=\"submit\" formmethod=\"%v\" formaction=\"%v\">%v</button>",
		html.EscapeString(l.method),
		html.EscapeString(l.url),
		html.EscapeString(l.name),
	)
}

// Notifier handles the notification delivery mechanism. It processes incoming notifications,
// stores them in the database, and sends them to the frontend for rendering.
type Notifier struct {
	channel  <-chan Notification
	upgrader websocket.Upgrader
	database *database.Database
}

// NewNotifier TODO: Restrict allowed origins to improve security by validating the 'Origin' header.
// NewNotifier initializes and returns a Notifier instance with a notification channel,
// WebSocket upgrader, and database connection.
func NewNotifier(db *database.Database) *Notifier {
	channel := make(chan Notification, 10)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &Notifier{
		channel:  channel,
		upgrader: upgrader,
		database: db,
	}
}

// notify processes a notification by formatting its content, saving it to the database,
// and sending it to the specified frontend host. It also logs errors encountered during the process.
func (n Notifier) notify(msg Notification) {
	// Format the notification content with the message and associated links.
	content := fmt.Sprintf("<span>%v</span>", msg.Message())
	content += "\n<form style='display: flex; flex-direction: column; gap: 2px; margin-top: 3px'>"
	for _, link := range msg.Links() {
		content += link.String()
	}
	content += "\n</form>"

	// Create a new message object with the notification content.
	message := &models.Message{
		Sender:   0,
		Receiver: 0,
		Content:  content,
		Created:  time.Now(),
	}

	// Send the notification to each target user and store it in the database.
	targets := msg.Targets()
	for _, t := range targets {
		message.Receiver = t
		_, err := n.database.Message.SendMessage(*message)
		if err != nil {
			log.Printf("could not insert notification message for %v: %v\n", t, err)
		}
	}

	// Prepare the notification payload for delivery to the frontend.
	payload := struct {
		Targets []int64         `json:"targets"`
		Message *models.Message `json:"message"`
	}{
		Targets: targets,
		Message: message,
	}

	// Encode the payload to JSON and send it to the frontend.
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)
	if err != nil {
		log.Println(err)
	}

	_, err = http.Post(fmt.Sprintf("http://%v:8080/notify", frontendHost), "", b)
	if err != nil {
		log.Printf("could not notify notification: %v\n", err)
	}
}

// userGetName returns the preferred display name of a user, prioritizing their nickname if available.
func userGetName(u *models.User) string {
	if len(u.Nickname) > 0 {
		return u.Nickname
	}
	return fmt.Sprintf("%v %v", u.FirstName, u.LastName)
}

// conditionalString returns the given string if the condition is true, or an empty string otherwise.
func conditionalString(b bool, s string) string {
	if b {
		return s
	}
	return ""
}

// getFrontendHost retrieves the frontend host address from environment variables,
// defaulting to "localhost" if the variable is not set.
func getFrontendHost() string {
	v := os.Getenv("FRONTEND_ADDRESS")
	if v == "" {
		v = "localhost"
	}
	return v
}
