package api

import (
	"encoding/json"
	"github.com/Pomog/sn/backend/models"
	"log"
	"net/http"
	"time"
)

// SendMessage handles the process of sending a message. It assigns the sender from the current session and stores the message in the database.
func SendMessage(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Create a new message instance
	message := models.Message{}

	// Decode the incoming JSON message data
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		log.Println(err)                           // Log the error
		writeStatusError(w, http.StatusBadRequest) // Respond with a bad request status
		return
	}

	// Assign the sender ID from the current session
	message.Sender = session.UserID

	// Save the message in the database and retrieve the new message ID
	id, err := Database.Message.SendMessage(message)
	panicIfErr(err)

	// Set additional message attributes
	message.MessageID = id
	message.Created = time.Now()

	// If the message is sent to a group, fetch sender data
	if message.IsGroup {
		u, err := Database.User.GetByID(message.Sender)
		panicIfErr(err)
		message.SenderData = u.Limited() // Assign limited user data to the sender
	}

	// Respond with the newly created message
	writeJSON(w, message)
}

// GetMessages retrieves messages for the current user based on the message details in the request.
func GetMessages(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	// Create a new message instance to hold the request data
	message := models.Message{}

	// Decode the incoming request for message details
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		log.Println(err)                           // Log the error
		writeStatusError(w, http.StatusBadRequest) // Respond with a bad request status
		return
	}

	// Assign the sender ID from the current session
	message.Sender = session.UserID

	// Retrieve messages from the database based on the message details
	messages, err := Database.Message.GetMessages(message)
	panicIfErr(err)

	// Respond with the list of retrieved messages
	writeJSON(w, messages)
}
