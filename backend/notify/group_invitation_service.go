package notify

import (
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"html"
)

// Invite represents the details of a group invitation, including the group and the target user.
type Invite struct {
	group  *models.Group
	target int64
}

// Invite method sends a group invitation to the specified user.
func (n Notifier) Invite(group *models.Group, target int64) {
	n.notify(Invite{
		group:  group,
		target: target,
	})
}

// Targets returns the list of recipients for the invitation notification.
func (n Invite) Targets() []int64 {
	return []int64{n.target} // Notify only the targeted user.
}

// Message generates the content of the invitation notification.
func (n Invite) Message() string {
	return fmt.Sprintf(
		"You have been invited to the group <strong>%v</strong>.",
		html.EscapeString(n.group.Name), // Escapes group name to ensure safe HTML rendering.
	)
}

// Links provides actions related to the group invitation, such as joining the group.
func (n Invite) Links() []Link {
	return []Link{
		{
			name:   "Join group",                                          // Action link for accepting the invitation.
			url:    fmt.Sprintf("/submit/group/%v/join", n.group.GroupID), // URL for joining the group.
			method: "POST",                                                // Method used to accept the invitation.
		},
	}
}
