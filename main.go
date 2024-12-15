// Package main is the entry point of the application, handling user commands via a CLI interface.
package main

// Import statements:
// - "flag": Used to parse command-line arguments for the CLI interface.
// - "fmt": Provides formatted I/O functions, like printing to the console.
// - "go-to-peer/file": Custom package for file chunking and reconstruction.
// - "go-to-peer/peer": Custom package for client-server communication.
// - "go-to-peer/util": Custom package for utility functions, such as logging.
import (
	"flag" // Command-line flag parsing library
	"fmt"  // Formatted I/O library
	"go-to-peer/peer"
	"go-to-peer/util" // Local utility package for logging and other reusable components
	//"os"
)

// main is the application's entry point.
// It initializes the logger, parses CLI arguments, and directs the user to appropriate functionality.
func main() {
	// Initialize the logger to ensure all events are logged with timestamps and file references.
	util.InitLogger()
	util.Logger.Println("Application started") // Log application start.

	// Define CLI commands:
	serverPort := flag.String("server", "", "Start a server on the specified port")
	peerAddress := flag.String("connect", "", "Connect to a peer at the specified address")
	listCatalog := flag.Bool("catalog", false, "List available files on the server")
	downloadFile := flag.String("download", "", "Download a file by name")

	// Parse the command-line arguments provided by the user.
	flag.Parse()

	// Start the server if the "server" flag is provided.
	if *serverPort != "" {
		peer.StartServer(*serverPort)
		return
	}

	// Connect to a peer if the "connect" flag is provided.
	if *peerAddress != "" {
		if *listCatalog {
			// List files available on the server.
			fmt.Println("Requesting file catalog from the server...")
			peer.RequestFileCatalog(*peerAddress)
		} else if *downloadFile != "" {
			// Download the specified file.
			fmt.Printf("Downloading file: %s\n", *downloadFile)
			err := peer.DownloadFile(*peerAddress, *downloadFile)
			if err != nil {
				fmt.Printf("Error downloading file %s: %v\n", *downloadFile, err)
				util.Logger.Printf("Error downloading file %s: %v", *downloadFile, err)
			} else {
				fmt.Printf("Successfully downloaded file: %s\n", *downloadFile)
			}
		} else {
			// If no valid action is provided, show usage.
			fmt.Println("Usage:")
			fmt.Println("  -catalog         : List available files on the server")
			fmt.Println("  -download <name> : Download a file by name")
		}
		return
	}

	// If no arguments are provided, show usage.
	fmt.Println("Usage:")
	fmt.Println("  -server <port>  : Start a server on the specified port")
	fmt.Println("  -connect <addr> : Connect to a peer at the specified address")
	fmt.Println("  -catalog        : List available files on the server")
	fmt.Println("  -download <name>: Download a file by name")
}
