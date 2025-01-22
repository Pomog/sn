package middleware

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/config"
	"Social_Network/pkg/models"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// HaveGroupAccess checks if the user is a member of the group and has access.
func HaveGroupAccess(ctx *socialnetwork.Context) {
	groupId := ctx.Request.URL.Query().Get("group_id")
	if groupId == "" {
		ctx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Group ID is required.",
		})
		return
	}
	userUUID := ctx.Values["userId"].(uuid.UUID)
	var mg = new(models.GroupMember)
	if err := mg.GetMember(ctx.Db.Conn, userUUID, uuid.MustParse(groupId), false); err != nil {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authorized.",
		})
		return
	}
	ctx.Values["role"] = mg.Role
	ctx.Values["member"] = mg
	ctx.Next()
}

// IsGroupAdmin checks if the user is an admin of the group.
func IsGroupAdmin(ctx *socialnetwork.Context) {
	role := ctx.Values["role"].(models.GroupMemberRole)
	if role != models.MemberRoleAdmin {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authorized.",
		})
		return
	}
	ctx.Next()
}

// CheckGroupRole checks if the user has the specified role in the group.
func CheckGroupRole(ctx *socialnetwork.Context, role models.GroupMemberRole) {
	_role, ok := ctx.Values["role"].(models.GroupMemberRole)
	if !ok {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authorized.",
		})
		return
	}
	if _role != role {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not authorized.",
		})
		return
	}
	ctx.Next()
}

const DirName = "uploads"

// ImageUploadMiddleware checks if the file is an image and uploads it.
func ImageUploadMiddleware(c *socialnetwork.Context) {
	// Parse the multipart form in the request
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing the form.",
		})
		return
	}
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error retrieving the file.",
		})
		return
	}
	defer file.Close()

	// Check if the file is an image
	ext := []string{".jpeg", ".jpg", ".png", ".svg+xml", ".gif"}
	if !contains(ext, strings.ToLower(filepath.Ext(handler.Filename))) {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "File type not allowed.",
		})
		return
	}

	id := uuid.New()
	pathImg := path.Join(DirName, id.String()+filepath.Ext(handler.Filename))

	// Create the file using the id as the name and the extension from the original file
	dst, err := os.Create(pathImg)
	if err != nil {
		c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "Error creating the file.",
		})
		return
	}

	// Write the file
	if _, err := io.Copy(dst, file); err != nil {
		c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "Error writing the file.",
		})
		return
	}
	c.Values["file"] = pathImg
	c.Next()
}

// contains checks if a string is in a list of strings.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// IsPostValid checks if the post data is valid.
func IsPostValid(c *socialnetwork.Context) {
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body.",
		})
		return
	}

	privacy := data["privacy"].(string)
	_, err := models.PostPrivacyFromString(privacy)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid privacy value.",
		})
		return
	}

	if data["content"] == nil || strings.TrimSpace(data["content"].(string)) == "" ||
		data["title"] == nil || strings.TrimSpace(data["title"].(string)) == "" ||
		(data["image_url"] == nil && strings.TrimSpace(data["image_url"].(string)) == "") {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data.",
		})
		return
	}
	c.Next()
}

// CreateGroupMiddleware checks if the group creation data is valid.
func CreateGroupMiddleware(c *socialnetwork.Context) {
	var token string
	headerBearer := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body.",
		})
		return
	}
	_, err := config.Sess.Start(c).Get(token)
	if err != nil {
		c.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not logged in.",
		})
		return
	}

	if data["title"] == nil || strings.TrimSpace(data["title"].(string)) == "" ||
		data["description"] == nil || strings.TrimSpace(data["description"].(string)) == "" ||
		(data["banner_url"] == nil && strings.TrimSpace(data["banner_url"].(string)) == "") {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data.",
		})
		return
	}
	c.Next()
}

// CreateEventMiddleware checks if the event creation data is valid.
func CreateEventMiddleware(c *socialnetwork.Context) {
	var token string
	headerBearer := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body.",
		})
		return
	}
	_, err := config.Sess.Start(c).Get(token)
	if err != nil {
		c.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not logged in.",
		})
		return
	}

	if data["group_id"] == nil || strings.TrimSpace(data["group_id"].(string)) == "" ||
		data["title"] == nil || strings.TrimSpace(data["title"].(string)) == "" ||
		data["description"] == nil || strings.TrimSpace(data["description"].(string)) == "" ||
		data["date_time"] == nil || strings.TrimSpace(data["date_time"].(string)) == "" {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data.",
		})
		return
	}
	c.Next()
}

func IsGroupPostValid(c *socialnetwork.Context) {
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		log.Println(err.Error())
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body",
		})
		return
	}
	if data["title"] == nil || strings.TrimSpace(data["title"].(string)) == "" ||
		data["content"] == nil || strings.TrimSpace(data["content"].(string)) == "" || data["privacy"] == nil || strings.TrimSpace(data["privacy"].(string)) == "" || (data["privacy"] != "public" && data["privacy"] != "private" && data["privacy"] != "group" && data["privacy"] != "almost private") {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data",
		})
		return
	}
	c.Next()
}
func IsGroupValid(c *socialnetwork.Context) {
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		log.Println(err.Error())
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body",
		})
		return
	}
	if data["title"] == nil || strings.TrimSpace(data["title"].(string)) == "" ||
		data["description"] == nil || strings.TrimSpace(data["description"].(string)) == "" {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data",
		})
		return
	}
	c.Next()
}
func IsGroupExist(c *socialnetwork.Context) {
	_groupId := c.Request.URL.Query().Get("group_id")
	group := new(models.Group)
	// Check if the group is uuid
	groupId, err := uuid.Parse(_groupId)

	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid group uuid",
		})
		return
	}
	if err := group.Get(c.Db.Conn, groupId, false, false); err != nil {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Group not found",
		})
		return
	}
	c.Values["group_id"] = groupId
	c.Values["group"] = group
	c.Next()
}
func IsInvitedUserExist(c *socialnetwork.Context) {
	_userId := c.Request.URL.Query().Get("user_id")
	fmt.Println("IsInvitedUserExist(c *socialnetwork.Context) {")
	fmt.Println("_userId")
	user := new(models.User)
	userId, err := uuid.Parse(_userId)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid user uuid",
		})
		return
	}
	if err := user.Get(c.Db.Conn, userId); err != nil {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "User not found",
		})
		return
	}
	// Check if the user is already in the group
	groupId := c.Values["group_id"].(uuid.UUID)
	member := new(models.GroupMember)
	if err := member.GetMember(c.Db.Conn, userId, groupId, false); err == nil &&
		(member.Status == models.MemberStatusInvited ||
			member.Status == models.MemberStatusAccepted) {
		c.Status(http.StatusConflict).JSON(map[string]string{
			"error": "User already in the group",
		})
		return
	}
	c.Values["invited_user_id"] = userId
	c.Next()
}
func NoGroupAccess(c *socialnetwork.Context) {
	groupId := c.Values["group_id"].(uuid.UUID)
	userUUID := c.Values["userId"].(uuid.UUID)
	var mg = new(models.GroupMember)
	if err := mg.GetMember(c.Db.Conn, userUUID, groupId, false); err == nil && mg.Status != models.MemberStatusDeclined {
		c.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are already a member of this group.",
		})
		return
	}
	c.Next()
}
func IsGroupPostExist(c *socialnetwork.Context) {
	groupId := c.Values["group_id"].(uuid.UUID)
	_postId := c.Request.URL.Query().Get("post_id")
	post := new(models.Post)
	postId, err := uuid.Parse(_postId)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid post uuid",
		})
		return
	}
	if err := post.Get(c.Db.Conn, postId); err != nil {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Post not found",
		})
		return
	}
	if post.GroupID != groupId {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Post not found",
		})
		return
	}
	c.Values["group_id"] = groupId
	c.Values["post_id"] = postId
	c.Next()
}
func IsInvitationExist(c *socialnetwork.Context) {
	_invitationId := c.Request.URL.Query().Get("invitation_id")
	member := new(models.GroupMember)
	invitationId, err := uuid.Parse(_invitationId)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid invitation uuid",
		})
		return
	}

	if err := member.GetMemberById(c.Db.Conn, invitationId, false); err != nil && member.Status != models.MemberStatusInvited {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Invitation not found",
		})
		return
	}
	c.Values["invitation_id"] = invitationId
	c.Values["member"] = member
	c.Next()
}
func IsAccessDemandExist(c *socialnetwork.Context) {
	_requestingId := c.Request.URL.Query().Get("requesting_id")
	member := new(models.GroupMember)
	requestingId, err := uuid.Parse(_requestingId)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid requesting uuid",
		})
		return
	}
	if err := member.GetMemberById(c.Db.Conn, requestingId, false); err != nil && member.Status != models.MemberStatusRequesting {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Requesting not found",
		})
		return
	}

	c.Values["requesting_id"] = requestingId
	c.Values["member"] = member
	c.Next()
}
func IsEventExist(c *socialnetwork.Context) {
	_eventId := c.Request.URL.Query().Get("event_id")
	event := new(models.Event)
	eventId, err := uuid.Parse(_eventId)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid event uuid",
		})
		return
	}
	if err := event.Get(c.Db.Conn, eventId, false, false); err != nil {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Event not found",
		})
		return
	}
	if event.DateTime.Before(time.Now()) {
		c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "Event is passed",
		})
		return
	}
	c.Values["event_id"] = eventId
	c.Values["event"] = event
	c.Next()
}

// CreateGroupMessageMiddleware is a middleware that checks if the data is valid
func CreateGroupMessageMiddleware(c *socialnetwork.Context) {
	var token string
	headerBearer := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(headerBearer, "Bearer ") {
		token = strings.TrimPrefix(headerBearer, "Bearer ")
	}
	var data = map[string]interface{}{}
	if err := c.BodyParser(&data); err != nil {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Error parsing request body",
		})
		return
	}
	_, err := config.Sess.Start(c).Get(token)
	if err != nil {
		c.Status(http.StatusUnauthorized).JSON(map[string]string{
			"error": "You are not logged in.",
		})
		return
	}
	if data["group_id"] == nil || strings.TrimSpace(data["group_id"].(string)) == "" ||
		data["content"] == nil || strings.TrimSpace(data["content"].(string)) == "" {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "Invalid data",
		})
		return
	}
	c.Next()
}
