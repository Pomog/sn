package notify

import (
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"html"
)

// Request structure holds details of the group and the user who requested to join.
type Request struct {
	group     *models.Group
	requester *models.User
}

// Request method triggers a notification for a group join request.
func (n Notifier) Request(group *models.Group, requester *models.User) {
	n.notify(Request{
		group:     group,
		requester: requester,
	})
}

// Targets identifies the intended recipients for the notification.
func (n Request) Targets() []int64 {
	return []int64{n.group.OwnerID} // Notify the group owner.
}

// Message generates a personalized message for the group owner about the join request.
func (n Request) Message() string {
	return fmt.Sprintf(
		"%v has requested to join your group <strong>%v</strong>",
		html.EscapeString(userGetName(n.requester)), // Securely escape user input to prevent HTML injection.
		html.EscapeString(n.group.Name),
	)
}

// Links provides actionable links for the recipient, such as viewing the requester's profile or accepting the request.
func (n Request) Links() []Link {
	return []Link{
		{
			name:   "Show profile",
			url:    fmt.Sprintf("/user/%v", n.requester.UserID), // Directs to the requester's profile.
			method: "GET",
		},
		{
			name:   "Accept request",
			url:    fmt.Sprintf("/submit/group/%v/invite/%v", n.group.GroupID, n.requester.UserID), // API to accept the request.
			method: "POST",
		},
	}
}
