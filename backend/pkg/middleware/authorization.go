package middleware

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/config"
	"Social_Network/pkg/models"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
