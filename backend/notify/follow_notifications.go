package notify

import (
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"html"
)

type Follow struct {
	follower  *models.User
	following int64
}

// Follow notification method for the Notifier.
// It sends a notification when a user follows another user.
func (n Notifier) Follow(follower *models.User, following int64) {
	// TODO: Log the follow event and ensure the notification reaches the correct user.
	n.notify(Follow{
		follower:  follower,
		following: following,
	})
}

// Targets returns the recipient(s) of the follow notification.
func (f Follow) Targets() []int64 {
	// Returns the user who was followed as the target of the notification.
	return []int64{f.following}
}

// Message generates the follow notification message.
func (f Follow) Message() string {
	// Constructs the notification message, indicating that the user has a new follower.
	return fmt.Sprintf("<strong>%v</strong> is now your follower!", html.EscapeString(userGetName(f.follower)))
}

// Links generates the links for the follow notification, including a profile link.
func (f Follow) Links() []Link {
	// Returns a link to the follower's profile page.
	return []Link{
		{
			name:   "See their profile",
			url:    fmt.Sprintf("/user/%v", f.follower.UserID),
			method: "GET",
		},
	}
}

type FollowAccepted struct {
	accepter *models.User
	target   int64
}

// FollowAccepted notification method for the Notifier.
// It sends a notification when a user accepts a follow request.
func (n Notifier) FollowAccepted(accepter *models.User, target int64) {
	// TODO: Track follow acceptance and update user relationships accordingly.
	n.notify(FollowAccepted{
		accepter: accepter,
		target:   target,
	})
}

// Targets returns the recipient(s) of the follow acceptance notification.
func (f FollowAccepted) Targets() []int64 {
	// The target is the user who was followed and accepted the request.
	return []int64{f.target}
}

// Message generates the follow acceptance notification message.
func (f FollowAccepted) Message() string {
	// Constructs the notification message, indicating that the user is now following another user.
	return fmt.Sprintf("You are now following <strong>%v</strong>!", html.EscapeString(userGetName(f.accepter)))
}

// Links generates the links for the follow acceptance notification, including a profile link.
func (f FollowAccepted) Links() []Link {
	// Returns a link to the acceptor's profile page.
	return []Link{
		{
			name:   "See their profile",
			url:    fmt.Sprintf("/user/%v", f.accepter.UserID),
			method: "GET",
		},
	}
}

type FollowRequest struct {
	requester *models.User
	target    int64
}

// FollowRequest notification method for the Notifier.
// It sends a notification when a user sends a follow request to another user.
func (n Notifier) FollowRequest(requester *models.User, target int64) {
	// TODO: Add validation for the requester's ability to send a follow request.
	n.notify(FollowRequest{
		requester: requester,
		target:    target,
	})
}

// Targets returns the recipient(s) of the follow request notification.
func (f FollowRequest) Targets() []int64 {
	// The target is the user who received the follow request.
	return []int64{f.target}
}

// Message generates the follow request notification message.
func (f FollowRequest) Message() string {
	// Constructs the notification message, indicating that a follow request was sent.
	return fmt.Sprintf("<strong>%v</strong> has sent you a follow request", html.EscapeString(userGetName(f.requester)))
}

// Links generates the links for the follow request notification, including a profile link.
func (f FollowRequest) Links() []Link {
	// Returns a link to the requester's profile page.
	return []Link{
		{
			name:   "See their profile",
			url:    fmt.Sprintf("/user/%v", f.requester.UserID),
			method: "GET",
		},
	}
}
