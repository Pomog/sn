package app

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HandlerFunc defines a type for request handlers
type HandlerFunc func(*Context)

// ErrorHandlerFunc defines a type for error handlers
type ErrorHandlerFunc func(*Context, int)

// Route represents a single route in the application
type Route struct {
	pattern  string
	handlers []HandlerFunc
	methods  map[string]bool
}

// App represents the core application structure
type App struct {
	Db               *db
	routes           []*Route
	onErrorCode      ErrorHandlerFunc
	globalMiddleware []HandlerFunc
}

// New creates and initializes a new App instance
func New() *App {
	return &App{
		routes:           make([]*Route, 0),
		globalMiddleware: make([]HandlerFunc, 0),
	}
}

// UseDb assigns a database connection to the app
func (app *App) UseDb(conn *sql.DB) {
	app.Db = &db{Conn: conn}
}

// handle registers a route with specific handlers and methods
func (app *App) handle(pattern string, handlers []HandlerFunc, methods ...string) {
	methodsMap := make(map[string]bool)
	for _, method := range methods {
		methodsMap[method] = true
	}
	handlers = append(app.globalMiddleware, handlers...)
	route := &Route{pattern: pattern, handlers: handlers, methods: methodsMap}
	app.routes = append(app.routes, route)
}

// Use adds global middleware to the app
func (app *App) Use(handlers ...HandlerFunc) {
	app.globalMiddleware = append(app.globalMiddleware, handlers...)
}

// Static serves static files from a specified directory
func (app *App) Static(path, dir string) {
	fileServer := http.FileServer(http.Dir(dir))
	app.GET(path+"*", func(c *Context) {
		http.StripPrefix(path, fileServer).ServeHTTP(c.ResponseWriter, c.Request)
	})
}

// Register HTTP methods (GET, POST, PUT, DELETE)
func (app *App) GET(path string, handler ...HandlerFunc) {
	app.handle(path, handler, "GET")
}

func (app *App) PUT(path string, handler ...HandlerFunc) {
	app.handle(path, handler, "PUT")
}

func (app *App) POST(path string, handler ...HandlerFunc) {
	app.handle(path, handler, "POST")
}

func (app *App) DELETE(path string, handler ...HandlerFunc) {
	app.handle(path, handler, "DELETE")
}

// OnErrorCode sets a custom handler for HTTP error codes
func (app *App) OnErrorCode(handler ErrorHandlerFunc) {
	app.onErrorCode = handler
}

// NotAllowed sends a 405 error response
func (app *App) NotAllowed(c *Context) {
	http.Error(c.ResponseWriter, "405 Method not allowed", http.StatusMethodNotAllowed)
}

// ServeHTTP handles incoming HTTP requests and routes them appropriately
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{ResponseWriter: w, Request: r, Db: app.Db, Values: make(map[any]any)}
	for _, route := range app.routes {
		if strings.HasSuffix(route.pattern, "*") {
			if strings.HasPrefix(r.URL.Path, strings.TrimSuffix(route.pattern, "*")) {
				if route.methods[r.Method] {
					c.handlers = route.handlers
					c.Next()
					return
				}
				app.NotAllowed(c)
				return
			}
		} else if r.URL.Path == route.pattern {
			if route.methods[r.Method] || r.Method == "OPTIONS" {
				c.handlers = route.handlers
				c.Next()
				return
			}
			app.NotAllowed(c)
			return
		}
	}
	if app.onErrorCode != nil {
		app.onErrorCode(c, http.StatusNotFound)
	} else {
		http.NotFound(w, r)
	}
}

var wg sync.WaitGroup

// checkServer verifies if the server is running
func checkServer(addr string) {
	resp, err := http.Get("http://" + addr)
	if err != nil {
		log.Println("Server is not running")
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Error closing response body: %v", err)
			}
		}(resp.Body)
		displayLaunchMessage(addr)
	}
}

// Run starts the application on the specified address
func (app *App) Run(addr string) error {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(addr, app); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Check if the server is running
	checkServer(addr)

	wg.Wait()
	return nil
}

// displayLaunchMessage prints a visually appealing launch message
func displayLaunchMessage(addr string) {
	fmt.Println("=============================================")
	fmt.Println("              🚀 Application Started         ")
	fmt.Println("=============================================")
	host, _ := os.Hostname()
	fmt.Printf("📌 Hostname       : %s\n", host)
	fmt.Printf("🌐 Listening Addr : http://%s\n", addr)
	fmt.Println("=============================================")
	fmt.Println("Ready to accept connections. Enjoy! 😊")
	fmt.Println("=============================================")
}
