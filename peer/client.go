// Package peer implements the client-side logic for connecting to other peers.
package peer

// Import statements:
// - "bufio": For buffered reading from a network connection.
// - "encoding/json": For encoding and decoding JSON messages.
// - "fmt": For user-facing messages.
// - "net": For establishing TCP connections.
// - "go-to-peer/util": For logging significant events.
import (
	"bufio"         // Buffered reading/writing to TCP connections.
	"encoding/json" // JSON encoding/decoding for structured message exchange.
	"fmt"           // Formatted I/O for user-facing messages.
	"go-to-peer/util"
	"net" // TCP networking for peer connections.
)

// ConnectToPeer establishes a connection to the specified peer address, exchanges messages,
// and logs significant events and errors. It uses a JSON-based communication protocol.
//
// Parameters:
// - address: The target peer's address in the format "host:port" (e.g., "127.0.0.1:8080").
//
// Behavior:
// - Sends a "HELLO" message to the peer.
// - Receives and processes the peer's response (e.g., metadata or other messages).
func ConnectToPeer(address string) {
	// Attempt to establish a TCP connection to the target peer.
	conn, err := net.Dial("tcp", address)
	if err != nil {
		// Log the connection failure.
		util.Logger.Printf("Failed to connect to peer at %s: %v", address, err)
		fmt.Printf("Failed to connect to peer at %s. Check logs for details.\n", address)
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close connection to %s: %v", address, closeErr)
		}
	}()

	// Log successful connection.
	util.Logger.Printf("Connected to peer at %s", address)
	fmt.Printf("Successfully connected to peer at %s\n", address)

	// Construct and send a "HELLO" message.
	msg := Message{
		Type:    "HELLO",
		Payload: "This is a test message from the client.",
	}
	data, err := EncodeMessage(msg)
	if err != nil {
		util.Logger.Printf("Failed to encode message: %v", err)
		fmt.Println("Error preparing message for peer. Check logs for details.")
		return
	}

	_, writeErr := conn.Write(append(data, '\n')) // Send JSON-encoded message with newline.
	if writeErr != nil {
		util.Logger.Printf("Failed to send message to peer at %s: %v", address, writeErr)
		fmt.Println("Error sending message to peer. Check logs for details.")
		return
	}

	// Log the successful message send.
	util.Logger.Printf("Sent HELLO message to peer at %s", address)

	// Read response from the peer.
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read response from peer at %s: %v", address, err)
		fmt.Println("Error receiving response from peer. Check logs for details.")
		return
	}

	// Decode the response into a Message structure.
	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode response from peer at %s: %v", address, decodeErr)
		fmt.Println("Error processing response from peer. Check logs for details.")
		return
	}

	// Handle the received message if it's of type "METADATA".
	if respMsg.Type == "METADATA" {
		var metadata Metadata
		metadataBytes, _ := json.Marshal(respMsg.Payload)
		_ = json.Unmarshal(metadataBytes, &metadata) // Parse payload into Metadata struct.

		// Log and display the received metadata.
		util.Logger.Printf("Received Metadata from peer at %s: %+v", address, metadata)
		fmt.Printf("Received Metadata from peer: %+v\n", metadata)
	}

	// Log and display the full response message for reference.
	util.Logger.Printf("Received response from peer at %s: %+v", address, respMsg)
	fmt.Printf("Received response from peer: %+v\n", respMsg)
}
