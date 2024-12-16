// Package main is the entry point of the application, handling user commands via a CLI interface.
package main

// Import statements:
// - "flag": Used to parse command-line arguments for the CLI interface.
// - "fmt": Provides formatted I/O functions, like printing to the console.
// - "go-to-peer/peer": Custom package for client-server communication.
// - "go-to-peer/util": Custom package for utility functions, such as logging.
import (
	"flag" // Command-line flag parsing library
	"fmt"  // Formatted I/O library
	"go-to-peer/peer"
	"go-to-peer/util" // Local utility package for logging and other reusable components
	"strings"
)

// main is the application's entry point.
// It initializes the logger, parses CLI arguments, and directs the user to appropriate functionality.
func main() {
	// Initialize the logger to ensure all events are logged with timestamps and file references.
	util.InitLogger()
	util.Logger.Println("Application started") // Log application start.

	// Define CLI commands:
	serverPort := flag.String("server", "", "Start a server on the specified port")
	peerAddresses := flag.String("connect", "", "Comma-separated list of peer addresses to connect to")
	listCatalog := flag.Bool("catalog", false, "List available files on all connected servers")
	fileHash := flag.String("download", "", "Download a file by its hash")
	fileName := flag.String("name", "", "Specify the original file name for the downloaded file")

	// Parse the command-line arguments provided by the user.
	flag.Parse()

	// Start the server if the "server" flag is provided.
	if *serverPort != "" {
		peer.StartServer(*serverPort)
		return
	}

	// Connect to peers if the "connect" flag is provided.
	if *peerAddresses != "" {
		addresses := splitCommaSeparated(*peerAddresses)

		if *listCatalog {
			// List files available on all connected servers.
			fmt.Println("Requesting file catalogs from all servers...")
			fileSources, err := peer.FetchFileCatalogs(addresses)
			if err != nil {
				fmt.Printf("Error fetching file catalogs: %v\n", err)
				util.Logger.Printf("Error fetching file catalogs: %v", err)
				return
			}

			// Display the combined catalog from all servers.
			fmt.Println("Combined File Catalog:")
			for hash, servers := range fileSources {
				fmt.Printf("- Hash: %s\n", hash)
				for _, server := range servers {
					fmt.Printf("  Available on server: %s\n", server)
				}
			}
			return
		}

		if *fileHash != "" {
			// Download the specified file by its hash.
			if *fileName == "" {
				fmt.Println("Error: Please specify the original file name using the -name flag.")
				return
			}

			fmt.Printf("Downloading file with hash: %s\n", *fileHash)
			err := peer.DownloadFileFromMultipleServers(*fileHash, *fileName, addresses)
			if err != nil {
				fmt.Printf("Error downloading file with hash %s: %v\n", *fileHash, err)
				util.Logger.Printf("Error downloading file with hash %s: %v", *fileHash, err)
			} else {
				fmt.Printf("Successfully downloaded file: %s\n", *fileName)
			}
			return
		}

		// If no valid action is provided, show usage.
		fmt.Println("Usage:")
		fmt.Println("  -catalog         : List available files on the servers")
		fmt.Println("  -download <hash> : Download a file by its hash (requires -name flag)")
		fmt.Println("  -name <name>     : Specify the original file name for the downloaded file")
		return
	}

	// If no arguments are provided, show usage.
	fmt.Println("Usage:")
	fmt.Println("  -server <port>   : Start a server on the specified port")
	fmt.Println("  -connect <addrs> : Connect to peer addresses (comma-separated)")
	fmt.Println("  -catalog         : List available files on the servers")
	fmt.Println("  -download <hash> : Download a file by its hash (requires -name flag)")
	fmt.Println("  -name <name>     : Specify the original file name for the downloaded file")
}

// splitCommaSeparated splits a comma-separated string into a slice of strings.
func splitCommaSeparated(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Split(input, ",")
}
