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
	"go-to-peer/file"
	"go-to-peer/util"
	"net" // TCP networking for peer connections.
	"os"
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
func ConnectToPeer(address string, requestedChunk string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		util.Logger.Printf("Failed to connect to peer at %s: %v", address, err)
		fmt.Printf("Failed to connect to peer at %s. Check logs for details.\n", address)
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close connection to %s: %v", address, closeErr)
		}
	}()

	util.Logger.Printf("Connected to peer at %s", address)
	fmt.Printf("Successfully connected to peer at %s\n", address)

	// Request a specific chunk.
	request := Message{
		Type: ChunkRequest,
		Payload: ChunkRequestPayload{
			ChunkID: requestedChunk,
		},
	}
	data, err := EncodeMessage(request)
	if err == nil {
		_, _ = conn.Write(append(data, '\n'))
		util.Logger.Printf("Requested chunk %s from peer at %s", requestedChunk, address)
	} else {
		util.Logger.Printf("Failed to encode CHUNK_REQUEST: %v", err)
		return
	}

	// Read the chunk response.
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read response from peer at %s: %v", address, err)
		return
	}

	// Decode and process the chunk response.
	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode response from peer at %s: %v", address, decodeErr)
		return
	}

	if respMsg.Type == ChunkResponse {
		var payload ChunkResponsePayload
		payloadBytes, _ := json.Marshal(respMsg.Payload)
		_ = json.Unmarshal(payloadBytes, &payload)

		// Verify the hash of the received chunk.
		if util.CalculateHash(payload.Data) == payload.Hash {
			util.Logger.Printf("Received and verified chunk %s from peer at %s", payload.ChunkID, address)
			fmt.Printf("Successfully received chunk %s\n", payload.ChunkID)
		} else {
			util.Logger.Printf("Chunk %s received from peer at %s failed integrity check", payload.ChunkID, address)
			fmt.Printf("Integrity check failed for chunk %s\n", payload.ChunkID)
		}
	}
}

// RequestFileCatalog requests the file catalog from the server.
//
// Parameters:
// - address: The server's address (e.g., "127.0.0.1:8080").
func RequestFileCatalog(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		util.Logger.Printf("Failed to connect to server at %s: %v", address, err)
		fmt.Printf("Failed to connect to server at %s. Check logs for details.\n", address)
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			util.Logger.Printf("Warning: Failed to close connection to server at %s: %v", address, closeErr)
		}
	}()

	// Send a FILE_CATALOG_REQUEST message.
	request := Message{
		Type: FileCatalogRequest,
	}
	data, err := EncodeMessage(request)
	if err == nil {
		_, _ = conn.Write(append(data, '\n'))
		util.Logger.Printf("Requested file catalog from server at %s", address)
	} else {
		util.Logger.Printf("Failed to encode FILE_CATALOG_REQUEST: %v", err)
		return
	}

	// Read the file catalog response.
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read file catalog response from server at %s: %v", address, err)
		return
	}

	// Decode and display the file catalog.
	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode file catalog response from server at %s: %v", address, decodeErr)
		return
	}
	if respMsg.Type == FileCatalogResponse {
		var catalog FileCatalog
		catalogBytes, _ := json.Marshal(respMsg.Payload)
		_ = json.Unmarshal(catalogBytes, &catalog)
		fmt.Printf("Received File Catalog:\n")
		for _, file := range catalog.Files {
			fmt.Printf("- %s (Size: %d bytes, Chunks: %d)\n", file.Name, file.Size, len(file.Chunks))
		}
		util.Logger.Printf("Successfully received and displayed file catalog from server at %s", address)
	}
}

func DownloadFile(address, fileName string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		util.Logger.Printf("Failed to connect to server at %s: %v", address, err)
		return err
	}
	defer conn.Close()

	fmt.Println("Requesting file catalog...")
	catalog, err := fetchCatalog(address)
	if err != nil {
		return fmt.Errorf("failed to fetch catalog: %w", err)
	}

	// Check if the requested file exists in the catalog.
	fileExists := false
	for _, file := range catalog.Files {
		if file.Name == fileName {
			fileExists = true
			break
		}
	}
	if !fileExists {
		return fmt.Errorf("file %s not found on server", fileName)
	}

	fmt.Printf("File %s is available. Starting download...\n", fileName)
	util.Logger.Printf("File %s is available for download", fileName)

	// Request file metadata.
	request := Message{
		Type: FileMetadataRequest,
		Payload: FileMetadataRequestPayload{
			FileName: fileName,
		},
	}
	data, err := EncodeMessage(request)
	if err != nil {
		util.Logger.Printf("Failed to encode FILE_METADATA_REQUEST: %v", err)
		return err
	}
	_, _ = conn.Write(append(data, '\n'))
	util.Logger.Printf("Requested metadata for file %s", fileName)

	// Read the metadata response.
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read file metadata response: %v", err)
		return err
	}

	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode metadata response: %v", decodeErr)
		return decodeErr
	}

	if respMsg.Type != FileMetadataResponse {
		util.Logger.Printf("Unexpected response type: %s", respMsg.Type)
		return fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}

	var metadata FileMetadataResponsePayload
	payloadBytes, _ := json.Marshal(respMsg.Payload)
	_ = json.Unmarshal(payloadBytes, &metadata)

	// Download each chunk and save it to the "chunks" directory.
	for _, chunk := range metadata.Chunks {
		chunkData, err := downloadChunk(conn, chunk)
		if err != nil {
			util.Logger.Printf("Failed to download chunk %s: %v", chunk, err)
			return err
		}
		err = saveChunk(chunk, chunkData)
		if err != nil {
			util.Logger.Printf("Failed to save chunk %s: %v", chunk, err)
			return err
		}
	}

	// Reconstruct the file.
	outputFile := fmt.Sprintf("downloads/%s", fileName)
	err = file.ReconstructFile(outputFile)
	if err != nil {
		util.Logger.Printf("Failed to reconstruct file %s: %v", fileName, err)
		return err
	}

	util.Logger.Printf("Successfully downloaded and reconstructed file: %s", fileName)
	return nil
}

// saveChunk saves the downloaded chunk to the "chunks" directory.
func saveChunk(chunkID string, data []byte) error {
	err := os.MkdirAll("chunks", 0755)
	if err != nil {
		return err
	}
	chunkPath := fmt.Sprintf("chunks/%s", chunkID)
	return os.WriteFile(chunkPath, data, 0644)
}

func downloadChunk(conn net.Conn, chunkID string) ([]byte, error) {
	// Send a CHUNK_REQUEST for the specified chunk.
	request := Message{
		Type: ChunkRequest,
		Payload: ChunkRequestPayload{
			ChunkID: chunkID,
		},
	}
	data, err := EncodeMessage(request)
	if err != nil {
		util.Logger.Printf("Failed to encode CHUNK_REQUEST: %v", err)
		return nil, err
	}
	_, _ = conn.Write(append(data, '\n'))
	util.Logger.Printf("Requested chunk %s", chunkID)

	// Read the CHUNK_RESPONSE.
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read CHUNK_RESPONSE for chunk %s: %v", chunkID, err)
		return nil, err
	}

	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode CHUNK_RESPONSE for chunk %s: %v", chunkID, decodeErr)
		return nil, decodeErr
	}

	if respMsg.Type != ChunkResponse {
		util.Logger.Printf("Unexpected response type: %s", respMsg.Type)
		return nil, fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}

	// Validate the chunk data and hash.
	var chunkPayload ChunkResponsePayload
	payloadBytes, _ := json.Marshal(respMsg.Payload)
	_ = json.Unmarshal(payloadBytes, &chunkPayload)

	if util.CalculateHash(chunkPayload.Data) != chunkPayload.Hash {
		util.Logger.Printf("Integrity check failed for chunk %s", chunkPayload.ChunkID)
		return nil, fmt.Errorf("integrity check failed for chunk %s", chunkPayload.ChunkID)
	}

	util.Logger.Printf("Successfully received and validated chunk %s", chunkPayload.ChunkID)
	return chunkPayload.Data, nil
}

// fetchCatalog retrieves the file catalog from the server.
func fetchCatalog(address string) (*FileCatalog, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		util.Logger.Printf("Failed to connect to server at %s: %v", address, err)
		return nil, err
	}
	defer conn.Close()

	request := Message{
		Type: FileCatalogRequest,
	}
	data, err := EncodeMessage(request)
	if err != nil {
		util.Logger.Printf("Failed to encode FILE_CATALOG_REQUEST: %v", err)
		return nil, err
	}
	_, _ = conn.Write(append(data, '\n'))
	util.Logger.Printf("Requested file catalog from server at %s", address)

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		util.Logger.Printf("Failed to read file catalog response: %v", err)
		return nil, err
	}

	respMsg, decodeErr := DecodeMessage([]byte(response))
	if decodeErr != nil {
		util.Logger.Printf("Failed to decode file catalog response: %v", decodeErr)
		return nil, decodeErr
	}

	if respMsg.Type != FileCatalogResponse {
		util.Logger.Printf("Unexpected response type: %s", respMsg.Type)
		return nil, fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}

	var catalog FileCatalog
	payloadBytes, _ := json.Marshal(respMsg.Payload)
	_ = json.Unmarshal(payloadBytes, &catalog)
	util.Logger.Printf("Received file catalog: %+v", catalog)
	return &catalog, nil
}
