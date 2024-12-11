// Package main is the entry point of the application, handling user commands via a CLI interface.
package main

// Import statements:
// - "flag": Used to parse command-line arguments for the CLI interface.
// - "fmt": Provides formatted I/O functions, like printing to the console.
// - "go-to-peer/file": Custom package for file chunking and reconstruction.
// - "go-to-peer/util": Custom package for utility functions, such as logging.
import (
	"flag" // Command-line flag parsing library
	"fmt"  // Formatted I/O library
	"go-to-peer/file"
	"go-to-peer/test"
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
	// -test-chunking: Used to test file chunking and reconstruction functionality.
	testChunking := flag.Bool("test-chunking", false, "Test file chunking and reconstruction")

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
	} else if *testChunking {
		fmt.Println("Testing file chunking and reconstruction...") // User-facing message
		util.Logger.Println("Testing chunking functions")          // Internal log

		// Generate test files of varying sizes.
		sizes := map[string]int64{
			"1KB":   1024,
			"100KB": 100 * 1024,
			"1MB":   1 * 1024 * 1024,
			"10MB":  10 * 1024 * 1024,
			"100MB": 100 * 1024 * 1024,
		}
		for label, size := range sizes {
			fmt.Printf("Generating %s test file...\n", label)
			util.Logger.Printf("Generating %s test file...", label)
			fileName, err := test.GenerateRandomFile(size)
			if err != nil {
				fmt.Printf("Failed to generate %s file. Check logs for details.\n", label)
				util.Logger.Printf("Failed to generate %s file: %v", label, err)
				continue
			}
			fmt.Printf("Generated %s test file: %s\n", label, fileName)
			util.Logger.Printf("Generated %s test file: %s", label, fileName)

			// Test chunking and reconstruction on the generated file.
			err = file.SplitFile(fileName)
			if err != nil {
				fmt.Printf("Error splitting %s file. Check logs for details.\n", label)
				util.Logger.Printf("Error splitting %s file: %v", label, err)
				continue
			}
			fmt.Printf("%s file successfully split into chunks.\n", label)
			util.Logger.Printf("%s file successfully split into chunks.", label)

			reconstructedFile := fmt.Sprintf("testdata/reconstructed_%s.dat", label)
			err = file.ReconstructFile(reconstructedFile)
			if err != nil {
				fmt.Printf("Error reconstructing %s file. Check logs for details.\n", label)
				util.Logger.Printf("Error reconstructing %s file: %v", label, err)
				continue
			}
			fmt.Printf("%s file successfully reconstructed: %s\n", label, reconstructedFile)
			util.Logger.Printf("%s file successfully reconstructed: %s", label, reconstructedFile)
		}

		// Optionally clean up the test data.
		fmt.Println("Cleaning up test data...")
		util.Logger.Println("Cleaning up test data...")
		err := test.DeleteTestDataDir()
		if err != nil {
			fmt.Println("Warning: Failed to clean up test data. Check logs for details.")
			util.Logger.Printf("Warning: Failed to clean up test data: %v", err)
		} else {
			fmt.Println("Test data cleaned up successfully.")
			util.Logger.Println("Test data cleaned up successfully.")
		}
	} else {
		// If no valid command is provided, display usage information.
		fmt.Println("Usage:")
		fmt.Println("  -upload <file_path> : Upload a file")
		fmt.Println("  -download <file_id> : Download a file")
		fmt.Println("  -test-chunking : Test file chunking and reconstruction")
		util.Logger.Println("User did not provide valid commands.")
	}
}
