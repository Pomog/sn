package api

import (
	"fmt"
	"github.com/Pomog/sn/backend/models"
	"github.com/Pomog/sn/backend/router"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

// FileUpload handles file uploads. Validates file size, extracts the file,
// and saves it using a unique token in the database.
func FileUpload(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size to prevent excessively large uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Println("Upload size exceeded the allowed limit.")
		http.Error(w, "File too large. Please upload a file smaller than 1MB.", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data. The form field name must be "file".
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to process file upload.", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			loggerFileHandler.Printf("Error while closing multipart.File: %v\n", err)
		}
	}(file)

	// Save the file to the database and generate a unique token for it.
	token, err := Database.File.Insert(file, fileHeader.Filename)
	if err != nil {
		log.Println("Error inserting file into the database:", err)
		http.Error(w, "File upload failed.", http.StatusInternalServerError)
		return
	}

	// Return a response containing the unique token for the uploaded file.
	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	writeJSON(w, response)
}

// FileDownload handles file downloads. Fetches file metadata using a token,
// retrieves the file, and serves it to the client.
func FileDownload(w http.ResponseWriter, r *http.Request) {
	// Extract the token from the request URL.
	token := router.GetSlug(r, 0)

	// Retrieve file metadata from the database using the token.
	fileData, err := Database.File.Get(token)
	if err != nil {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}

	// Open the physical file from the upload directory.
	filePath := path.Join(models.UploadsPath, fileData.Token+fileData.Extension)
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		http.Error(w, "Unable to retrieve the file.", http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			loggerFileHandler.Printf("Error while closing multipart.File: %v\n", err)
		}
	}(file)

	// Determine the appropriate content type based on the file extension.
	var contentType string
	switch fileData.Extension {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	default:
		contentType = "application/octet-stream"
	}

	// Set response headers to indicate the file type and suggest a filename.
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileData.Name))

	// Stream the file contents to the response.
	if _, err = io.Copy(w, file); err != nil {
		log.Println("Error writing file to response:", err)
		http.Error(w, "Error streaming the file.", http.StatusInternalServerError)
	}
}
