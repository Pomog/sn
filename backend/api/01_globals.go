package api

import (
	"github.com/Pomog/sn/backend/database"
	"github.com/Pomog/sn/backend/notify"
	"log"
	"os"
	"time"
)

// Shared variables for the API package
var (
	Database          *database.Database
	Notify            *notify.Notifier
	loggerFileHandler = log.New(os.Stdout, "FileHandler: ", log.LstdFlags)
)

// Shared constants
const sessionDuration = 24 * time.Hour
const maxUploadSize = 1024 * 1024 // Maximum file size allowed for uploads: 1MB
