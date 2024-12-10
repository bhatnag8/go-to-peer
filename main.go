// Package main is the entry point of the application, handling user commands via a CLI interface.
package main

// Import statements:
// - "flag": Used to parse command-line arguments for the CLI interface.
// - "fmt": Provides formatted I/O functions, like printing to the console.
// - "go-to-peer/util": Custom package for utility functions, such as logging.
import (
	"flag"            // Command-line flag parsing library
	"fmt"             // Formatted I/O library
	"go-to-peer/util" // Local utility package for logging and other reusable components
)

// main is the application's entry point.
// It initializes the logger, parses CLI arguments, and directs the user to appropriate functionality.
func main() {
	// Initialize the logger to ensure all events are logged with timestamps and file references.
	util.InitLogger()
	util.Logger.Println("Application started") // Log application start.

	// Define CLI commands:
	// -upload: Used to specify the path of a file to upload.
	uploadCmd := flag.String("upload", "", "Path to the file to upload")
	// -download: Used to specify the ID of a file to download.
	downloadCmd := flag.String("download", "", "File ID to download")

	// Parse the command-line arguments provided by the user.
	flag.Parse()

	// Handle the CLI commands based on user input.
	if *uploadCmd != "" {
		fmt.Printf("Uploading file: %s\n", *uploadCmd)
		util.Logger.Printf("User requested to upload file: %s", *uploadCmd)
		// Placeholder for file upload logic (to be implemented in later milestones).
	} else if *downloadCmd != "" {
		fmt.Printf("Downloading file with ID: %s\n", *downloadCmd)
		util.Logger.Printf("User requested to download file with ID: %s", *downloadCmd)
		// Placeholder for file download logic (to be implemented in later milestones).
	} else {
		// If no valid command is provided, display usage information.
		fmt.Println("Usage:")
		fmt.Println("  -upload <file_path> : Upload a file")
		fmt.Println("  -download <file_id> : Download a file")
		util.Logger.Println("User did not provide valid commands.")
	}
}
