package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"net/http"
)
import _ "github.com/mattn/go-sqlite3"

// Initializes the SQLite database connection and applies migrations.
func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on") // Enables foreign key constraints.
	if err != nil {
		panic(err)
	}

	// Applies any outstanding migrations to update the schema.
	if err := applyMigrations(db); err != nil {
		panic(err)
	}

	return db
}

// Embeds all migration files from the "migrations" directory.
//
//go:embed migrations/*
var migrations embed.FS

// Handles database migrations using embedded migration files.
func applyMigrations(db *sql.DB) error {
	// Creates a migration source from embedded files.
	sourceInstance, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return fmt.Errorf("unable to create migration source: %w", err)
	}

	// Configures the SQLite database as the migration target.
	targetInstance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to configure SQLite target: %w", err)
	}

	// Sets up the migration manager with the source and target.
	m, err := migrate.NewWithInstance("httpfs", sourceInstance, "sqlite", targetInstance)
	if err != nil {
		return fmt.Errorf("failed to initialize migration manager: %w", err)
	}

	// Runs migrations to update the database schema.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration error: %w", err)
	}

	// Cleans up resources associated with the migration source.
	return sourceInstance.Close()
}
