// Package peer implements the server-side logic for listening to peer connections.
package peer

// Import statements:
// - "bufio": For buffered reading from a connection.
// - "fmt": For user-facing messages (e.g., server status).
// - "net": For TCP networking.
// - "os": For error handling and logging.
// - "go-to-peer/util": For logging significant events.
import (
	"bufio"           // Buffered reading/writing to TCP connections.
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
func handleConnection(conn net.Conn) {
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close connection to peer %s: %v", conn.RemoteAddr(), closeErr)
		}
	}()

	// Log and print the connection establishment.
	peerAddr := conn.RemoteAddr().String()
	util.Logger.Printf("Connected to peer: %s", peerAddr)
	fmt.Printf("Peer connected: %s\n", peerAddr)

	// Dynamically generate metadata for the peer connection.
	metadata := Metadata{
		PeerID:    generatePeerID(), // Generate a unique PeerID (dynamic logic here).
		Hostname:  conn.LocalAddr().String(),
		ChunkList: getAvailableChunks(), // Replace with actual logic for fetching available chunks.
	}

	// Construct and send the metadata message.
	message := Message{
		Type:    "METADATA",
		Payload: metadata,
	}
	data, err := EncodeMessage(message)
	if err == nil {
		_, _ = conn.Write(append(data, '\n')) // Send JSON-encoded metadata.
		util.Logger.Printf("Sent metadata to peer %s: %+v", peerAddr, metadata)
	} else {
		util.Logger.Printf("Failed to encode or send metadata to peer %s: %v", peerAddr, err)
	}

	// Read and handle incoming messages from the peer.
	reader := bufio.NewReader(conn)
	for {
		// Read a message from the peer.
		message, err := reader.ReadString('\n')
		if err != nil {
			util.Logger.Printf("Connection closed by peer %s: %v", peerAddr, err)
			fmt.Printf("Peer disconnected: %s\n", peerAddr)
			return
		}

		// Decode the received message.
		msg, decodeErr := DecodeMessage([]byte(message))
		if decodeErr != nil {
			util.Logger.Printf("Failed to decode message from peer %s: %v", peerAddr, decodeErr)
			continue
		}

		// Log and display the received message.
		util.Logger.Printf("Received message from peer %s: %+v", peerAddr, msg)
		fmt.Printf("Message from peer %s: %+v\n", peerAddr, msg)
	}
}

// generatePeerID generates a unique identifier for the server.
// Replace this with a proper implementation (e.g., UUID or other mechanisms).
func generatePeerID() string {
	return "server-123" // Placeholder for now.
}

// getAvailableChunks returns a list of available file chunks on the server.
// Replace this with logic to fetch actual chunks.
func getAvailableChunks() []string {
	return []string{"chunk_0", "chunk_1"} // Placeholder logic.
}
