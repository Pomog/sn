package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migration performs the database migration based on the specified action.
func Migration(DB *sql.DB, migration *Migrations) { // Pass by reference
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Define the migration directory and database path
	migrationDir := currentDir + "/pkg/db/migrations/sqlite/"
	databasePath := currentDir + "/pkg/db/sqlite/social-network.db"

	// Initialize the migration instance
	m, err := migrate.New("file://"+migrationDir, "sqlite://"+databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(m)

	// Perform the migration based on the action specified
	switch strings.ToLower(migration.Action) {
	case "-up":
		// Apply one migration (1 Up)
		if err := m.Steps(1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Migration Error: ", err)
		}
	case "-down":
		// Rollback one migration (1 Down)
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Migration Error: ", err)
		}
	case "-up--all":
		// Apply all migrations (Up)
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Migration Error: ", err)
		}
	case "-down--all":
		// Rollback all migrations (Down)
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Migration Error: ", err)
		}
	case "-to":
		// Migrate directly to the target version
		if err := m.Migrate(uint(migration.Version)); err != nil {
			fmt.Println("Migration Error: ", err)
		}
	}

	// Check the current version and dirty state of the database
	currentVersion, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Fatal(err)
	}
	fmt.Println("Current database version:", currentVersion)
	fmt.Println("Dirty state of Database: ", dirty)
}
