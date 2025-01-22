package main

import (
	socialnetwork "Social_Network/app"
	"Social_Network/app/middleware/cors"
	"Social_Network/pkg/config"
	"Social_Network/pkg/db/sqlite"
	"Social_Network/pkg/handlers"
	"Social_Network/pkg/middleware"
	"Social_Network/pkg/tools"

	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load environment variables from .env file
	err := tools.LoadEnv(".env")
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Ensure necessary directories exist
	ensureDirectory(middleware.DirName)

	// Parse command-line arguments for migration options
	migrate := parseMigrationArgs(os.Args[1:])

	// Initialize the backend application
	app := initializeApp()

	// Configure and set up the database
	database := sqlite.OpenDB(&migrate)
	app.UseDb(database)

	// Add middleware for CORS and static file serving
	configureMiddleware(app)

	// Register all application handlers
	handlers.HandleAll(app)

	// Set up session management
	config.Sess.UseDB(app.Db.Conn)

	// Start the application server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	log.Printf("Starting server on port: %s\n", port)
	if err := app.Run(":" + port); err != nil {
		panic(err)
	}
}

// ensureDirectory checks if a directory exists and creates it if not.
func ensureDirectory(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err := os.MkdirAll(dirName, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dirName, err)
		}
		log.Printf("Directory created: %s\n", dirName)
	}
}

// parseMigrationArgs parses command-line arguments for database migrations.
func parseMigrationArgs(args []string) sqlite.Migrations {
	migrate := sqlite.Migrations{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-up") || strings.HasPrefix(arg, "-down") || strings.HasPrefix(arg, "-to") {
			migrate.Migration = true
			parts := strings.Split(arg, "=")
			if len(parts) == 2 {
				version, err := strconv.Atoi(parts[1])
				if err != nil || version <= 0 {
					log.Println("Invalid migration version")
				} else {
					migrate.Target = true
					migrate.Version = version
					migrate.Action = parts[0]
				}
			} else {
				migrate.Target = true
				migrate.Action = arg
			}
		} else {
			migrate.Migration = false
		}
	}
	return migrate
}

// initializeApp creates and initializes the application instance.
func initializeApp() *socialnetwork.App {
	return socialnetwork.New()
}

// configureMiddleware sets up CORS and static file serving middleware.
func configureMiddleware(app *socialnetwork.App) {
	app.Use(cors.New(cors.Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		AllowCredentials: true,
		ExposedHeaders:   []string{},
		MaxAge:           86400,
	}))
	app.Static("/uploads", middleware.DirName)
}
