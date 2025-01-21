package tools

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv loads environment variables from a specified file path
// The file should have key-value pairs in the format "KEY=VALUE" on each line.
func LoadEnv(filePath string) error {
	// Open the file containing environment variables
	file, err := os.Open(filePath)
	if err != nil {
		return err // Return the error if the file cannot be opened
	}
	defer file.Close() // Ensure the file is closed when the function returns

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Split the line into key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip lines that do not have a valid "KEY=VALUE" format
		}

		// Set the environment variable with the parsed key and value
		key, value := parts[0], parts[1]
		os.Setenv(key, value)
	}

	// Check if an error occurred during scanning
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil // Return nil if no errors occurred
}
