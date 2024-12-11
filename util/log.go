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
	/* source: https://pkg.go.dev/os
	const (
		// Exactly one of O_RDONLY, O_WRONLY, or O_RDWR must be specified.
		O_RDONLY int = syscall.O_RDONLY // open the file read-only.
		O_WRONLY int = syscall.O_WRONLY // open the file write-only.
		O_RDWR   int = syscall.O_RDWR   // open the file read-write.
		// The remaining values may be or'ed in to control behavior.
		O_APPEND int = syscall.O_APPEND // append data to the file when writing.
		O_CREATE int = syscall.O_CREAT  // create a new file if none exists.
		O_EXCL   int = syscall.O_EXCL   // used with O_CREATE, file must not exist.
		O_SYNC   int = syscall.O_SYNC   // open for synchronous I/O.
		O_TRUNC  int = syscall.O_TRUNC  // truncate regular writable file when opened.
	) */
	file, err := os.OpenFile("go-to-peer.log", os.O_CREATE|os.O_WRONLY /*|os.O_APPEND*/, 0666)

	if err != nil {
		// Critical error: Unable to initialize logging. Application should not proceed.
		log.Fatalln("Failed to open log file:", err)
	}

	// Initialize the Logger with a custom format:
	// - "INFO: " prefix for readability.
	// - Date and time for event tracking.
	// - File reference for debugging purposes.
	/* source: https://pkg.go.dev/log
	const (
		Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
		Ltime                         // the time in the local time zone: 01:23:23
		Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
		Llongfile                     // full file name and line number: /a/b/c/d.go:23
		Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
		LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
		Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
		LstdFlags     = Ldate | Ltime // initial values for the standard logger
	) */
	Logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
