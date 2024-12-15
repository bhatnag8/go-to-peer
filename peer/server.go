// Package peer implements the server-side logic for listening to peer connections.
package peer

// Import statements:
// - "bufio": For buffered reading from a connection.
// - "fmt": For user-facing messages (e.g., server status).
// - "net": For TCP networking.
// - "os": For error handling and logging.
// - "go-to-peer/util": For logging significant events.
import (
	"bufio" // Buffered reading/writing to TCP connections.
	"encoding/json"
	"path/filepath"
	"strings"

	//"encoding/json"
	"fmt"             // Formatted I/O for user-facing messages.
	"go-to-peer/util" // Logging utility for significant events.
	"net"             // TCP networking for peer connections.
	"os"              // OS-level functions for error handling and logging.
)

// StartServer starts a TCP server to listen for incoming peer connections.
//
// Parameters:
// - port: The port on which the server will listen for incoming connections.
//
// Behavior:
// - Listens on the specified port for incoming connections.
// - Handles each connection in a separate goroutine to support concurrent peers.
func StartServer(port string) {
	// Start the TCP listener on the specified port.
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		// Log the startup failure and terminate the application.
		util.Logger.Printf("Error starting server on port %s: %v", port, err)
		fmt.Printf("Error: Unable to start server on port %s. Check logs for details.\n", port)
		os.Exit(1)
	}
	defer func() {
		if closeErr := listener.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close listener on port %s: %v", port, closeErr)
		}
	}()

	// Log and print the server startup status.
	util.Logger.Printf("Server listening on port %s", port)
	fmt.Printf("Server successfully started on port %s...\n", port)

	// Accept incoming connections in a loop.
	for {
		conn, err := listener.Accept()
		if err != nil {
			util.Logger.Printf("Failed to accept connection: %v", err)
			fmt.Println("Error: Failed to accept a connection. Check logs for details.")
			continue
		}
		// Handle the connection in a separate goroutine for concurrency.
		go handleConnection(conn)
	}
}

// handleConnection handles an incoming peer connection.
//
// Parameters:
// - conn: The network connection established with the peer.
//
// Behavior:
// - Sends metadata to the connected peer.
// - Reads incoming messages and logs or processes them as necessary.
// Add a new message type for file catalog requests and responses.
const (
	FileCatalogRequest  = "FILE_CATALOG_REQUEST"
	FileCatalogResponse = "FILE_CATALOG_RESPONSE"
)

// Updated handleConnection to handle catalog requests.
func handleConnection(conn net.Conn) {
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close connection to peer %s: %v", conn.RemoteAddr(), closeErr)
		}
	}()

	peerAddr := conn.RemoteAddr().String()
	util.Logger.Printf("Connected to peer: %s", peerAddr)
	fmt.Printf("Peer connected: %s\n", peerAddr)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			util.Logger.Printf("Connection closed by peer %s: %v", peerAddr, err)
			fmt.Printf("Peer disconnected: %s\n", peerAddr)
			return
		}

		// Decode the message.
		msg, decodeErr := DecodeMessage([]byte(message))
		if decodeErr != nil {
			util.Logger.Printf("Failed to decode message from peer %s: %v", peerAddr, decodeErr)
			continue
		}

		// Handle different message types.
		switch msg.Type {

		// Handle FILE_CATALOG_REQUEST messages.
		case FileCatalogRequest:
			// Generate the file catalog dynamically.
			catalog, err := createCatalog("server_files") // Directory with server files.
			if err != nil {
				util.Logger.Printf("Failed to generate file catalog: %v", err)
				continue
			}

			// Send the file catalog to the client.
			response := Message{
				Type:    FileCatalogResponse,
				Payload: catalog,
			}
			data, encodeErr := EncodeMessage(response)
			if encodeErr == nil {
				_, _ = conn.Write(append(data, '\n'))
				util.Logger.Printf("Sent file catalog to peer %s", peerAddr)
			} else {
				util.Logger.Printf("Failed to encode FILE_CATALOG_RESPONSE: %v", encodeErr)
			}

		// Handle FILE_METADATA_REQUEST messages.
		case FileMetadataRequest:
			var payload FileMetadataRequestPayload
			payloadBytes, _ := json.Marshal(msg.Payload)
			_ = json.Unmarshal(payloadBytes, &payload)

			// Get the catalog and find the requested file.
			catalog, err := createCatalog("server_files")
			if err != nil {
				util.Logger.Printf("Failed to load catalog: %v", err)
				continue
			}

			var responsePayload FileMetadataResponsePayload
			for _, file := range catalog.Files {
				if file.Name == payload.FileName {
					responsePayload = FileMetadataResponsePayload{
						FileName: file.Name,
						Chunks:   file.Chunks,
					}
					break
				}
			}

			// Send the file metadata to the client.
			response := Message{
				Type:    FileMetadataResponse,
				Payload: responsePayload,
			}
			data, encodeErr := EncodeMessage(response)
			if encodeErr == nil {
				_, _ = conn.Write(append(data, '\n'))
				util.Logger.Printf("Sent metadata for file %s to peer %s", payload.FileName, peerAddr)
			} else {
				util.Logger.Printf("Failed to encode FILE_METADATA_RESPONSE: %v", encodeErr)
			}

		// Add logic for other message types (e.g., CHUNK_REQUEST) here as needed.
		case ChunkRequest:
			var payload ChunkRequestPayload
			payloadBytes, _ := json.Marshal(msg.Payload)
			_ = json.Unmarshal(payloadBytes, &payload)

			// Find the file name for the requested chunk.
			fileName := "" // Placeholder for the actual file name
			catalog, err := createCatalog("server_files")
			if err == nil {
				for _, file := range catalog.Files {
					for _, chunk := range file.Chunks {
						if chunk == payload.ChunkID {
							fileName = file.Name
							break
						}
					}
					if fileName != "" {
						break
					}
				}
			}

			if fileName == "" {
				util.Logger.Printf("Failed to find file for chunk %s", payload.ChunkID)
				continue
			}

			// Retrieve the chunk data.
			chunkData, chunkHash, err := getChunkData(payload.ChunkID, fileName)
			if err != nil {
				util.Logger.Printf("Failed to retrieve chunk %s for peer %s: %v", payload.ChunkID, peerAddr, err)
				continue
			}

			// Respond with the chunk data.
			response := Message{
				Type: ChunkResponse,
				Payload: ChunkResponsePayload{
					ChunkID: payload.ChunkID,
					Data:    chunkData,
					Hash:    chunkHash,
				},
			}
			data, encodeErr := EncodeMessage(response)
			if encodeErr == nil {
				_, _ = conn.Write(append(data, '\n'))
				util.Logger.Printf("Sent chunk %s to peer %s", payload.ChunkID, peerAddr)
			} else {
				util.Logger.Printf("Failed to encode CHUNK_RESPONSE for chunk %s: %v", payload.ChunkID, encodeErr)
			}

		default:
			util.Logger.Printf("Received unknown message type from peer %s: %s", peerAddr, msg.Type)
		}
	}
}

func getChunkData(chunkID string, fileName string) ([]byte, string, error) {
	filePrefix := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	chunkFilePath := filepath.Join("chunks", filePrefix, chunkID)

	data, err := os.ReadFile(chunkFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read chunk %s: %w", chunkID, err)
	}

	hash := util.CalculateHash(data)
	return data, hash, nil
}
