// Package peer defines the communication protocol and peer connectivity logic.
package peer

// Import statements:
// - "encoding/json": For JSON serialization and deserialization of messages.
// - "fmt": For formatted error messages.
// - "go-to-peer/util": For logging significant events.
import (
	"encoding/json" // JSON encoding/decoding for structured message exchange.
	"fmt"           // Formatted I/O for error handling.
	"go-to-peer/util"
)

// Message represents the structure of messages exchanged between peers.
//
// Fields:
// - Type: The type of the message (e.g., "HELLO", "METADATA").
// - Payload: The actual data being sent, which varies depending on the message type.
type Message struct {
	Type    string      `json:"type"`    // Type of the message (e.g., "HELLO", "METADATA").
	Payload interface{} `json:"payload"` // The actual data being sent.
}

// Metadata represents the structure of metadata exchanged between peers.
//
// Fields:
// - PeerID: A unique identifier for the peer.
// - Hostname: The hostname or address of the peer.
// - ChunkList: A list of available chunks on the peer.
type Metadata struct {
	PeerID    string   `json:"peer_id"`    // Unique identifier for the peer.
	Hostname  string   `json:"hostname"`   // Peer hostname or address.
	ChunkList []string `json:"chunk_list"` // List of available chunks.
}

// EncodeMessage converts a Message struct into a JSON byte array.
//
// Parameters:
// - msg: The Message struct to encode.
//
// Returns:
// - []byte: The JSON-encoded message.
// - error: An error object if the encoding fails.
//
// Logging:
// - Logs an error message if encoding fails.
func EncodeMessage(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		// Log the error and return a wrapped error.
		util.Logger.Printf("Error encoding message: %v", err)
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}
	return data, nil
}

// DecodeMessage parses a JSON byte array into a Message struct.
//
// Parameters:
// - data: The JSON byte array to decode.
//
// Returns:
// - Message: The decoded Message struct.
// - error: An error object if the decoding fails.
//
// Logging:
// - Logs an error message if decoding fails.
func DecodeMessage(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		// Log the error and return a wrapped error.
		util.Logger.Printf("Error decoding message: %v", err)
		return Message{}, fmt.Errorf("failed to decode message: %w", err)
	}
	return msg, nil
}

// Add new message types for chunk transfers.
const (
	ChunkRequest  = "CHUNK_REQUEST"  // Message type for requesting a file chunk.
	ChunkResponse = "CHUNK_RESPONSE" // Message type for sending a requested chunk.
)

// ChunkRequestPayload represents the payload structure for chunk requests.
type ChunkRequestPayload struct {
	ChunkID string `json:"chunk_id"` // ID of the requested chunk.
}

// ChunkResponsePayload represents the payload structure for chunk responses.
type ChunkResponsePayload struct {
	ChunkID string `json:"chunk_id"` // ID of the chunk being sent.
	Data    []byte `json:"data"`     // Actual chunk data.
	Hash    string `json:"hash"`     // Hash of the chunk data for integrity verification.
}

const (
	FileMetadataRequest  = "FILE_METADATA_REQUEST"  // Request metadata for a specific file.
	FileMetadataResponse = "FILE_METADATA_RESPONSE" // Response containing file metadata.
)

// FileMetadataRequestPayload represents the payload for metadata requests.
type FileMetadataRequestPayload struct {
	FileName string `json:"file_name"` // The name of the file being requested.
}

// FileMetadataResponsePayload represents the payload for metadata responses.
type FileMetadataResponsePayload struct {
	FileName string   `json:"file_name"` // The name of the file.
	Chunks   []string `json:"chunks"`    // List of chunk IDs for the file.
}
