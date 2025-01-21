package handlers

import (
	socialnetwork "Social_Network/app"
	"Social_Network/pkg/config"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/models"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/mail"
)

type userCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates the provided credentials.
func (c *userCredentials) Validate() error {
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.New("invalid email")
	}

	if c.Password == "" {
		return errors.New("password is missing. please provide it")
	}

	return nil
}

var loginHandler = func(ctx *socialnetwork.Context) {
	var credentials = userCredentials{}

	if err := ctx.BodyParser(&credentials); err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while parsing the form data.",
			"status":  http.StatusInternalServerError,
			"data":    nil,
		})
		return
	}

	if err := credentials.Validate(); err != nil {
		ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"session": "",
			"message": err.Error(),
			"status":  http.StatusBadRequest,
			"data":    nil,
		})
		return
	}

	newUser := models.User{
		Email:    credentials.Email,
		Password: credentials.Password,
	}

	err := newUser.Get(ctx.Db.Conn, credentials.Email, true)
	if err != nil {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
			"session": "",
			"message": "invalid email.",
			"status":  http.StatusUnauthorized,
			"data":    nil,
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(newUser.Password), []byte(credentials.Password)); err != nil {
		ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
			"session": "",
			"message": "Invalid credentials. Please try again.",
			"status":  http.StatusUnauthorized,
			"data":    nil,
		})
		return
	}

	idSession, err := config.Sess.Start(ctx).Set(newUser.ID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while starting the session.",
			"status":  http.StatusInternalServerError,
			"data":    nil,
		})
		return
	}
	ctx.JSON(map[string]interface{}{
		"session": idSession,
		"message": "User successfully logged.",
		"status":  "200",
		"data":    newUser,
	})
}

var loginRoute = route{
	method: http.MethodPost,
	path:   "/login",
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.NoAuthRequired,
		loginHandler,
	},
}

var registrationHandler = func(ctx *socialnetwork.Context) {
	var newUser = models.User{}
	if err := ctx.BodyParser(&newUser); err != nil {
		ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while parsing the form data.",
			"status":  http.StatusBadRequest,
			"data":    nil,
		})
		return
	}

	err := newUser.Validate(ctx.Db.Conn)
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"session": "",
			"message": err.Error(),
			"status":  http.StatusBadRequest,
			"data":    nil,
		})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while hashing the password.",
			"status":  http.StatusInternalServerError,
			"data":    nil,
		})
		return
	}
	newUser.Password = string(newHash)
	if err := newUser.Create(ctx.Db.Conn); err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while creating the user. User already registered.",
			"status":  http.StatusInternalServerError,
			"data":    nil,
		})
		return
	}

	idSession, err := config.Sess.Start(ctx).Set(newUser.ID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"session": "",
			"message": "Error while starting the session.",
			"status":  http.StatusInternalServerError,
			"data":    nil,
		})
		return
	}
	ctx.JSON(map[string]interface{}{
		"session": idSession,
		"message": "User successfully registered and logged.",
		"status":  "200",
		"data":    newUser,
	})
}

var registrationRoute = route{
	method: http.MethodPost,
	path:   "/registration",
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.NoAuthRequired,
		registrationHandler,
	},
}

var healthHandler = func(ctx *socialnetwork.Context) {
	ctx.WriteString("💻Server is Ok!")
}

var healthRoute = route{
	method: http.MethodGet,
	path:   "/health",
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		healthHandler,
	},
}

func LogoutHandler(ctx *socialnetwork.Context) {
	token := ctx.Values["token"].(string)
	err := config.Sess.Start(ctx).Delete(token)
	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "Error while deleting session",
		})
		return
	}

	ctx.Status(http.StatusOK)
}

var logoutRoute = route{
	method: http.MethodDelete,
	path:   "/logout",
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,
		LogoutHandler,
	},
}

func meHandler(ctx *socialnetwork.Context) {
	userId := ctx.Values["userId"].(uuid.UUID)
	user := models.User{}
	err := user.Get(ctx.Db.Conn, userId)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK).JSON(user)
}

var meRoute = route{
	method: http.MethodGet,
	path:   "/me",
	middlewareAndHandler: []socialnetwork.HandlerFunc{
		middleware.AuthRequired,
		meHandler,
	},
}

func init() {
	AllHandler[loginRoute.path] = loginRoute
	AllHandler[logoutRoute.path] = logoutRoute
	AllHandler[meRoute.path] = meRoute
	AllHandler[registrationRoute.path] = registrationRoute
	AllHandler[healthRoute.path] = healthRoute
}
