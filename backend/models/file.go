package models

import (
	"database/sql"
	"fmt"
	"github.com/Pomog/sn/backend/queries"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"time"
)

var loggerFileModel = log.New(os.Stdout, "FileModel: ", log.LstdFlags)

// PersistPath constant for file storage paths
const PersistPath = "./persist"              // Root directory for persistent storage
const UploadsPath = PersistPath + "/uploads" // Directory for uploaded files

// File represents metadata about a stored file.
type File struct {
	Token     string    `json:"UUID"`      // Unique identifier for the file
	Name      string    `json:"name"`      // Original file name
	Extension string    `json:"extension"` // File extension
	Created   time.Time `json:"created"`   // Timestamp when the file was created
}

// pointerSlice provides pointers to the struct fields for database scanning.
func (x *File) pointerSlice() []interface{} {
	return []interface{}{
		&x.Token,
		&x.Name,
		&x.Extension,
		&x.Created,
	}
}

// FileModel encapsulates database and query logic for file operations.
type FileModel struct {
	queries queries.QueryProvider // Query provider for prepared SQL queries
	db      *sql.DB               // Database connection
}

func MakeFileModel(db *sql.DB) FileModel {
	// Ensure the uploads directory exists. os.MkdirAll creates parent directories if needed.
	_ = os.Mkdir(UploadsPath, os.ModePerm)

	return FileModel{
		queries: queries.NewQueryProvider(db, "file"),
		db:      db,
	}
}

func (model FileModel) Get(token string) (*File, error) {
	stmt := model.queries.Prepare("get")

	row := stmt.QueryRow(token)

	file := &File{}
	err := row.Scan(file.pointerSlice()...)
	if err != nil {
		return nil, fmt.Errorf("File/Get: %w", err)
	}

	return file, nil
}

// extensionRegex Regex to extract file extension
var extensionRegex = regexp.MustCompile(`\.\w+$`)

// Insert saves a new file to the filesystem and database.
func (model FileModel) Insert(file io.Reader, filename string) (string, error) {
	stmt := model.queries.Prepare("insert")

	extension := extensionRegex.FindString(filename)

	token := uuid.New().String()

	out, err := os.Create(path.Join(UploadsPath, token+extension))
	if err != nil {
		return "", fmt.Errorf("File/Insert: %w", err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			loggerFileModel.Printf("Error while closing file: %v\n", err)
		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("File/Insert: %w", err)
	}

	// Insert metadata into the database
	_, err = stmt.Exec(token, filename, extension)
	if err != nil {
		return "", fmt.Errorf("File/Insert: %w", err)
	}

	return token, err
}

// Delete removes a file from the filesystem and database.
func (model FileModel) Delete(token string) (bool, error) {
	stmt := model.queries.Prepare("delete")

	file, err := model.Get(token)
	if err != nil {
		return false, fmt.Errorf("File/Delete: %w", err)
	}

	res, err := stmt.Exec(token)
	if err != nil {
		return false, fmt.Errorf("File/Delete: %w", err)
	}

	// Remove the file from the filesystem
	err = os.Remove(path.Join(UploadsPath, token+file.Extension))
	if err != nil {
		return false, fmt.Errorf("File/Delete: %w", err)
	}

	n, _ := res.RowsAffected()

	return n > 0, nil
}
