package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FollowerStatus defines the status of a follower.
type FollowerStatus string

// Followers is a slice of Follower.
type Followers []Follower

// Constants for different follower statuses.
const (
	StatusRequested FollowerStatus = "requested"
	StatusAccepted  FollowerStatus = "accepted"
	StatusDeclined  FollowerStatus = "declined"
)

// Follower represents a follower relationship between two users.
type Follower struct {
	ID         uuid.UUID      `sql:"type:uuid;primary key"` // Unique ID for the follower entry.
	FollowerID uuid.UUID      `sql:"type:uuid"`             // ID of the follower.
	FolloweeID uuid.UUID      `sql:"type:uuid"`             // ID of the user being followed.
	Status     FollowerStatus // Status of the follower (requested, accepted, declined).
	CreatedAt  time.Time      // Timestamp when the follower was created.
	UpdatedAt  time.Time      // Timestamp when the follower was last updated.
}

// Create a new follower in the database.
func (follower *Follower) Create(db *sql.DB) error {
	// Set default values for the follower entry.
	follower.ID = uuid.New()        // Generate a new UUID for the follower.
	follower.CreatedAt = time.Now() // Set the creation time.
	follower.UpdatedAt = time.Now() // Set the update time.

	// SQL query to insert a new follower.
	query := `INSERT INTO followers (id, follower_id, followee_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	// Prepare the SQL statement.
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close() // Ensure the statement is closed after execution.

	// Execute the SQL statement to insert the new follower.
	_, err = stmt.Exec(
		follower.ID,
		follower.FollowerID,
		follower.FolloweeID,
		follower.Status,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// Get a follower by follower_id and followee_id.
func (follower *Follower) Get(db *sql.DB, reverse ...bool) error {
	// Optionally reverse the follower and followee IDs.
	if len(reverse) > 0 && reverse[0] {
		follower.FollowerID, follower.FolloweeID = follower.FolloweeID, follower.FollowerID
	}

	// SQL query to fetch a follower.
	query := `SELECT id, follower_id, followee_id, status, created_at, updated_at FROM followers WHERE follower_id = $1 AND followee_id = $2 AND deleted_at IS NULL`

	// Execute the query and map the result to the follower struct.
	err := db.QueryRow(query, follower.FollowerID, follower.FolloweeID).Scan(
		&follower.ID,
		&follower.FollowerID,
		&follower.FolloweeID,
		&follower.Status,
		&follower.CreatedAt,
		&follower.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// Update the status of an existing follower.
func (follower *Follower) Update(db *sql.DB) error {
	// SQL query to update the status of a follower.
	query := `UPDATE followers SET status = $1, updated_at = $2 WHERE id = $3`

	// Prepare the SQL statement.
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close() // Ensure the statement is closed after execution.

	// Execute the SQL statement to update the follower's status.
	_, err = stmt.Exec(follower.Status, time.Now(), follower.ID)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// Delete (soft-delete) a follower by setting deleted_at to the current time.
func (follower *Follower) Delete(db *sql.DB) error {
	// SQL query to soft-delete the follower by updating deleted_at.
	query := `UPDATE followers SET deleted_at = $1 WHERE id = $2`

	// Prepare the SQL statement.
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close() // Ensure the statement is closed after execution.

	// Execute the SQL statement to mark the follower as deleted.
	_, err = stmt.Exec(time.Now(), follower.ID)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// GetAllByFolloweeID Get all accepted followers for a given followee ID.
func (followers *Followers) GetAllByFolloweeID(db *sql.DB, followeeID uuid.UUID) error {
	// SQL query to fetch all accepted followers for the given followee.
	query := `SELECT id, follower_id, followee_id, status, created_at, updated_at FROM followers WHERE followee_id = $1 AND status= "accepted" AND  deleted_at IS NULL`

	// Execute the query.
	rows, err := db.Query(query, followeeID)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close() // Ensure rows are closed after use.

	// Iterate through the rows and scan into the followers slice.
	for rows.Next() {
		var follower Follower
		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FolloweeID,
			&follower.Status,
			&follower.CreatedAt,
			&follower.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		*followers = append(*followers, follower)
	}

	return nil
}

// CountAllByFolloweeID Count the total number of accepted followers for a given followee ID.
func (followers *Followers) CountAllByFolloweeID(db *sql.DB, followeeID uuid.UUID) int {
	// SQL query to count the accepted followers for the given followee.
	query := `SELECT COUNT(id) FROM followers WHERE followee_id = $1 AND status = $2 AND deleted_at IS NULL`
	var count int

	// Execute the query and get the count.
	err := db.QueryRow(query, followeeID, StatusAccepted).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// GetAllByFollowerID Get all accepted followers for a given follower ID.
func (followers *Followers) GetAllByFollowerID(db *sql.DB, followerID uuid.UUID) error {
	// SQL query to fetch all accepted followers for the given follower.
	query := `SELECT id, follower_id, followee_id, status, created_at, updated_at FROM followers WHERE follower_id = $1 AND status= "accepted" AND deleted_at IS NULL`

	// Execute the query.
	rows, err := db.Query(query, followerID)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close() // Ensure rows are closed after use.

	// Iterate through the rows and scan into the followers slice.
	for rows.Next() {
		var follower Follower
		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FolloweeID,
			&follower.Status,
			&follower.CreatedAt,
			&follower.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		*followers = append(*followers, follower)
	}

	return nil
}

// CountAllByFollowerID Count the total number of accepted followers for a given follower ID.
func (followers *Followers) CountAllByFollowerID(db *sql.DB, followerID uuid.UUID) int {
	// SQL query to count the accepted followers for the given follower.
	query := `SELECT COUNT(id) FROM followers WHERE follower_id = $1 AND status = $2 AND deleted_at IS NULL`
	var count int

	// Execute the query and get the count.
	err := db.QueryRow(query, followerID, StatusAccepted).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
