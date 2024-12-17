// Package main is the entry point of the application, handling user commands via a CLI interface.
package main

import (
	"flag" // Command-line flag parsing library
	"fmt"  // Formatted I/O library
	"go-to-peer/peer"
	"go-to-peer/util" // Local utility package for logging and other reusable components
	"runtime"         // For performance monitoring (CPU usage)
	//"runtime/debug"   // To collect garbage before measuring performance
	"strings"
	"time" // For measuring execution time
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

	// Collect start time and memory stats for performance measurement.
	startTime := time.Now()
	var startMemStats runtime.MemStats
	runtime.ReadMemStats(&startMemStats)

	// Start the server if the "server" flag is provided.
	if *serverPort != "" {
		peer.StartServer(*serverPort)
		measurePerformance(startTime, startMemStats)
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
				measurePerformance(startTime, startMemStats)
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
			measurePerformance(startTime, startMemStats)
			return
		}

		if *fileHash != "" {
			// Download the specified file by its hash.
			if *fileName == "" {
				fmt.Println("Error: Please specify the original file name using the -name flag.")
				measurePerformance(startTime, startMemStats)
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
			measurePerformance(startTime, startMemStats)
			return
		}

		// If no valid action is provided, show usage.
		fmt.Println("Usage:")
		fmt.Println("  -catalog         : List available files on the servers")
		fmt.Println("  -download <hash> : Download a file by its hash (requires -name flag)")
		fmt.Println("  -name <name>     : Specify the original file name for the downloaded file")
		measurePerformance(startTime, startMemStats)
		return
	}

	// If no arguments are provided, show usage.
	fmt.Println("Usage:")
	fmt.Println("  -server <port>   : Start a server on the specified port")
	fmt.Println("  -connect <addrs> : Connect to peer addresses (comma-separated)")
	fmt.Println("  -catalog         : List available files on the servers")
	fmt.Println("  -download <hash> : Download a file by its hash (requires -name flag)")
	fmt.Println("  -name <name>     : Specify the original file name for the downloaded file")
	measurePerformance(startTime, startMemStats)
}

// splitCommaSeparated splits a comma-separated string into a slice of strings.
func splitCommaSeparated(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Split(input, ",")
}

// measurePerformance calculates and prints the performance metrics of the program.
func measurePerformance(startTime time.Time, startMemStats runtime.MemStats) {
	// Collect end time and memory stats.
	endTime := time.Now()
	var endMemStats runtime.MemStats
	runtime.ReadMemStats(&endMemStats)

	// Calculate elapsed time.
	elapsedTime := endTime.Sub(startTime)

	// Calculate CPU usage (this is an approximation; for detailed stats, use profiling tools like pprof).
	cpuPercentage := 100.0 * float64(runtime.NumGoroutine()) / float64(runtime.NumCPU())

	// Calculate memory usage.
	usedMemory := endMemStats.Alloc - startMemStats.Alloc

	// Print performance metrics.
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("Execution Time: %v\n", elapsedTime)
	fmt.Printf("Memory Usage: %d bytes\n", usedMemory)
	fmt.Printf("Approximate CPU Usage: %.2f%%\n", cpuPercentage)
}
