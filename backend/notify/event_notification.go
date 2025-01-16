package notify

import (
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"html"
)

type EventCreated struct {
	group   *models.Group
	event   *models.Event
	creator *models.User
	members []*models.User
}

func (n Notifier) EventCreated(
	group *models.Group,
	event *models.Event,
	creator *models.User,
	members []*models.User,
) {
	// TODO: Implement logging or validation to ensure the correct members receive the notification.
	n.notify(EventCreated{
		group:   group,
		event:   event,
		creator: creator,
		members: members,
	})
}

func (n EventCreated) Targets() []int64 {
	ids := make([]int64, 0, len(n.members)-1)
	// TODO: Filter out users who have opted out of receiving event notifications.
	for _, member := range n.members {
		if member.UserID != n.creator.UserID {
			ids = append(ids, member.UserID)
		}
	}
	return ids
}

func (n EventCreated) Message() string {
	// TODO: Customize the message content based on user preferences, such as language or notification type.
	return fmt.Sprintf(
		"Event <strong>%v</strong> has been created in %v by %v",
		html.EscapeString(n.event.Title),
		html.EscapeString(n.group.Name),
		html.EscapeString(userGetName(n.creator)),
	)
}

func (n EventCreated) Links() []Link {
	// TODO: Add conditional logic for including additional links, such as event updates or invites.
	return []Link{
		{
			name:   "Show event",
			url:    fmt.Sprintf("/event/%v", n.event.EventID),
			method: "GET",
		},
	}
}
