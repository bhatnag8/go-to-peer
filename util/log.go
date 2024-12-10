// Package util provides common utility functions used across the application.
// This file contains a logging utility to ensure all significant events are recorded for debugging and traceability.
package util

// Import statements:
// - "log": Provides logging functionality to write messages to a file or console.
// - "os": Enables operations on the operating system, such as file creation.
import (
	"log" // Standard logging package
	"os"  // OS-level functions, such as file handling
)

// Logger is a global variable that represents the application's logger instance.
// It is used to record log messages with consistent formatting across the application.
var Logger *log.Logger

// InitLogger initializes the global Logger instance.
// It creates or appends to a log file ("go-to-peer.log") and sets the logging format.
// SIL 4 compliance ensures every log entry has a timestamp, severity level, and source reference.
func InitLogger() {
	// Open or create the log file with write and append permissions.
	file, err := os.OpenFile("go-to-peer.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Critical error: Unable to initialize logging. Application should not proceed.
		log.Fatalln("Failed to open log file:", err)
	}

	// Initialize the Logger with a custom format:
	// - "INFO: " prefix for readability.
	// - Date and time for event tracking.
	// - File reference for debugging purposes.
	Logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
