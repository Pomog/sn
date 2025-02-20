package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strings"

	"html"
	"time"

	"github.com/google/uuid"
)

type Users []User

type User struct {
	ID    uuid.UUID `sql:"type:uuid;primary key" json:"id"`
	Email string    `sql:"type:varchar(100);unique" json:"email"`
	// Pseudo      string    `sql:"type:uuid;unique" json:"pseudo"`
	Password    string       `sql:"type:varchar(100)" json:"password"`
	FirstName   string       `sql:"type:varchar(100)" json:"firstName"`
	LastName    string       `sql:"type:varchar(100)" json:"lastName"`
	DateOfBirth time.Time    `json:"dateOfBirth"`
	AvatarImage string       `sql:"type:varchar(255)" json:"avatarImage"`
	Nickname    string       `sql:"type:varchar(100);unique" json:"nickname"`
	AboutMe     string       `sql:"type:text" json:"aboutMe"`
	IsPublic    bool         `json:"isPublic"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	DeletedAt   sql.NullTime `json:"deletedAt"`
}

func (u *User) Validate(db *sql.DB) error {
	requiredFields := []string{"Email", "Password", "FirstName", "LastName", "DateOfBirth"}

	v := reflect.ValueOf(u).Elem()

	for _, field := range requiredFields {
		f := v.FieldByName(field)
		if f.Interface() == reflect.Zero(f.Type()).Interface() {
			return errors.New(strings.ToLower(field) + " is missing. Please provide it.")
		}
	}

	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("Invalid email")
	}

	if len(u.Password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}

	if u.AvatarImage == "" {
		u.AvatarImage = "uploads/default-avatar.png"
	}

	if u.Nickname == "" {
		u.Nickname = uuid.NewString()
	} else {
		query := `SELECT COUNT(*) FROM users WHERE nickname = $1`
		var count int
		row := db.QueryRow(query, u.Nickname)
		err := row.Scan(&count)
		if err != nil {
			return fmt.Errorf("unable to query from database: ", err)
		}

		if count > 0 {
			return errors.New("Nickname already used.")
		}
	}

	return nil
}

// Create a new user
func (u *User) Create(db *sql.DB) error {
	// Mux.Lock()
	// defer Mux.Unlock()
	// Define the user default properties
	u.ID = uuid.New()
	// user.Pseudo = uuid.NewString()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	// if user.Nickname == "" {
	// 	user.Nickname = uuid.NewString()
	// }
	query := `INSERT INTO users (id, email, password, first_name, last_name, date_of_birth, avatar_image, nickname, about_me, is_public, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	stmt, err := db.Prepare(query)
	if err != nil {

		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		u.ID.String(),
		html.EscapeString(u.Email),
		// user.Pseudo,
		u.Password,
		html.EscapeString(u.FirstName),
		html.EscapeString(u.LastName),
		u.DateOfBirth,
		html.EscapeString(u.AvatarImage),
		html.EscapeString(u.Nickname),
		html.EscapeString(u.AboutMe),
		u.IsPublic,
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// Get a user by its ID
func (u *User) Get(db *sql.DB, identifier interface{}, password ...bool) error {
	if identifier == "" {
		fmt.Println("identifier cannot be an empty string")
		return errors.New("identifier cannot be an empty string")
	}
	// Mux.RLock()
	// defer Mux.RUnlock()
	query := `SELECT id, email, password, first_name, last_name, date_of_birth, avatar_image, nickname, about_me, is_public, created_at, updated_at FROM users WHERE id=$1 OR email=$1 OR nickname=$1`
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	defer stmt.Close()
	switch id := identifier.(type) {
	case string:
		if id == "" {
			return errors.New("identifier cannot be an empty string")
		}

		err := stmt.QueryRow(identifier).Scan(
			&u.ID,
			&u.Email,
			// &user.Pseudo,
			&u.Password,
			&u.FirstName,
			&u.LastName,
			&u.DateOfBirth,
			&u.AvatarImage,
			&u.Nickname,
			&u.AboutMe,
			&u.IsPublic,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		if (len(password) > 0 && password[0] == false) || len(password) == 0 {
			u.Password = ""
		}
	case uuid.UUID:
		if id == uuid.Nil {
			fmt.Println("identifier cannot be nil UUID")
			return errors.New("identifier cannot be nil UUID")
		}

		err := db.QueryRow(query, identifier).Scan(
			&u.ID,
			&u.Email,
			// &user.Pseudo,
			&u.Password,
			&u.FirstName,
			&u.LastName,
			&u.DateOfBirth,
			&u.AvatarImage,
			&u.Nickname,
			&u.AboutMe,
			&u.IsPublic,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		if len(password) > 0 && password[0] {
			u.Password = ""
		}
	default:
		fmt.Println("identifier type not supported")
		return errors.New("identifier type not supported")
	}

	fmt.Println("identifier ********************")
	fmt.Println(identifier)
	fmt.Println("USER ********************")
	fmt.Println(u)
	return nil
}

// Update a user
func (u *User) Update(db *sql.DB) error {
	// Mux.Lock()
	// defer Mux.Unlock()
	query := `UPDATE users SET email=$1, password=$2, first_name=$3, last_name=$4, date_of_birth=$5, avatar_image=$6, nickname=$7, about_me=$8, is_public=$9, updated_at=$10 WHERE id=$11`

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		html.EscapeString(u.Email),
		html.EscapeString(u.Password),
		html.EscapeString(u.FirstName),
		html.EscapeString(u.LastName),
		u.DateOfBirth,
		html.EscapeString(u.AvatarImage),
		html.EscapeString(u.Nickname),
		html.EscapeString(u.AboutMe),
		u.IsPublic,
		time.Now(),
		u.ID,
		// user.Pseudo,
	)

	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}

	return nil
}

// Delete a user
func (u *User) Delete(db *sql.DB) error {
	query := `UPDATE users SET deleted_at=$1 WHERE id=$2`

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare the query. %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), u.ID)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	return nil
}

// GetAll users
func (users *Users) GetAll(db *sql.DB) error {
	query := `SELECT id, email, password, first_name, last_name, date_of_birth, avatar_image, nickname, about_me, is_public, created_at, updated_at FROM users WHERE deleted_at IS NULL`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			// &user.Pseudo,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.AvatarImage,
			&user.Nickname,
			&user.AboutMe,
			&user.IsPublic,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		*users = append(*users, user)
	}

	return nil
}
func (users *Users) GetFlow(db *sql.DB, userid uuid.UUID) error {
	query := `
	SELECT DISTINCT u.*
	FROM users u
	JOIN followers f ON (u.id = f.follower_id OR u.id = f.followee_id)
	WHERE f.status = 'accepted' -- Vous pouvez ajouter des conditions supplémentaires ici si nécessaire
	AND (f.follower_id = $1 OR f.followee_id = $1);`

	rows, err := db.Query(query, userid)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.AvatarImage,
			&user.Nickname,
			&user.AboutMe,
			&user.IsPublic,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("unable to execute the query. %v", err)
		}
		*users = append(*users, user)
	}

	return nil
}
