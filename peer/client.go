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
	"strings"
	"sync"

	//"go-to-peer/file"
	"go-to-peer/util"
	"net" // TCP networking for peer connections.
	"os"
	//"strings"
	//"sync"
)

// DownloadFileFromMultipleServers downloads a file using its hash from multiple servers.
func DownloadFileFromMultipleServers(fileHash string, fileName string, servers []string) error {
	// Fetch metadata for the file from one of the servers.
	catalog, err := fetchCatalog(servers[0])
	if err != nil {
		return fmt.Errorf("failed to fetch catalog from server: %w", err)
	}

	var fileChunks []string
	for _, file := range catalog.Files {
		if file.Hash == fileHash {
			fileChunks = file.Chunks
			break
		}
	}
	if len(fileChunks) == 0 {
		return fmt.Errorf("file with hash %s not found on servers", fileHash)
	}

	// Distribute chunks across servers in a round-robin manner.
	chunkToServer := make(map[string]string)
	for i, chunk := range fileChunks {
		chunkToServer[chunk] = servers[i%len(servers)]
	}

	// Display progress to the user.
	progress := make(chan string, len(fileChunks))
	defer close(progress)
	go func() {
		for msg := range progress {
			fmt.Println(msg)
		}
	}()

	// Download chunks in parallel.
	chunkQueue := make(chan string, len(fileChunks))
	errChan := make(chan error, len(fileChunks))
	var wg sync.WaitGroup

	for chunk, server := range chunkToServer {
		chunkQueue <- fmt.Sprintf("%s|%s", server, chunk) // Encode server and chunk.
	}
	close(chunkQueue)

	numWorkers := 10 // Adjust as needed.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range chunkQueue {
				parts := strings.Split(job, "|")
				server, chunk := parts[0], parts[1]
				err := downloadChunkFromServer(server, chunk, fileHash, progress)
				if err != nil {
					errChan <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error during chunk download: %w", err)
		}
	}

	// Reconstruct the file after all chunks are downloaded.
	outputDir := "downloads"
	err = file.ReconstructFile(outputDir, fileHash)
	if err != nil {
		fmt.Printf("Failed to reconstruct file: %v\n", err)
	}

	util.Logger.Printf("Successfully downloaded and reconstructed file: %s", fileName)
	return nil
}

func RequestFileCatalog(servers []string) {
	for _, address := range servers {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			util.Logger.Printf("Failed to connect to server at %s: %v", address, err)
			fmt.Printf("Failed to connect to server at %s. Check logs for details.\n", address)
			continue
		}
		defer conn.Close()

		// Send a FILE_CATALOG_REQUEST message.
		request := Message{Type: FileCatalogRequest}
		data, err := EncodeMessage(request)
		if err != nil {
			util.Logger.Printf("Failed to encode FILE_CATALOG_REQUEST: %v", err)
			continue
		}
		_, _ = conn.Write(append(data, '\n'))
		util.Logger.Printf("Requested file catalog from server at %s", address)

		// Read the file catalog response.
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			util.Logger.Printf("Failed to read file catalog response from server at %s: %v", address, err)
			continue
		}

		respMsg, decodeErr := DecodeMessage([]byte(response))
		if decodeErr != nil {
			util.Logger.Printf("Failed to decode file catalog response from server at %s: %v", address, decodeErr)
			continue
		}

		if respMsg.Type != FileCatalogResponse {
			util.Logger.Printf("Unexpected response type: %s", respMsg.Type)
			fmt.Printf("Unexpected response type: %s\n", respMsg.Type)
			continue
		}

		var catalog FileCatalog
		payloadBytes, _ := json.Marshal(respMsg.Payload)
		_ = json.Unmarshal(payloadBytes, &catalog)

		// Display the catalog for the server.
		fmt.Printf("Received File Catalog from server %s:\n", address)
		for _, file := range catalog.Files {
			fmt.Printf("- %s (Size: %d bytes, Chunks: %d, Hash: %s)\n", file.Name, file.Size, len(file.Chunks), file.Hash)
		}
		util.Logger.Printf("Successfully received and displayed file catalog from server %s", address)
	}
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

// FetchFileCatalogs fetches catalogs from multiple servers and maps file hashes to servers.
func FetchFileCatalogs(servers []string) (map[string][]string, error) {
	fileSources := make(map[string][]string) // Map: file hash -> list of servers.

	for _, server := range servers {
		catalog, err := fetchCatalog(server)
		if err != nil {
			util.Logger.Printf("Failed to fetch catalog from %s: %v", server, err)
			continue
		}

		for _, file := range catalog.Files {
			if _, exists := fileSources[file.Hash]; !exists {
				fileSources[file.Hash] = []string{}
			}
			fileSources[file.Hash] = append(fileSources[file.Hash], server)
		}
	}

	return fileSources, nil
}

func fetchCatalog(address string) (*FileCatalog, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		util.Logger.Printf("Failed to connect to server at %s: %v", address, err)
		return nil, err
	}
	defer conn.Close()

	request := Message{Type: FileCatalogRequest}
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
	util.Logger.Printf("Received file catalog from %s: %+v", address, catalog)
	return &catalog, nil
}

func downloadChunkFromServer(server string, chunkID string, fileHash string, progress chan<- string) error {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %w", server, err)
	}
	defer conn.Close()

	chunkData, err := downloadChunk(conn, chunkID)
	if err != nil {
		return fmt.Errorf("failed to download chunk %s from server %s: %w", chunkID, server, err)
	}

	err = saveChunk(chunkID, fileHash, chunkData)
	if err != nil {
		return fmt.Errorf("failed to save chunk %s: %w", chunkID, err)
	}

	progress <- fmt.Sprintf("Downloaded chunk %s from server %s", chunkID, server)
	return nil
}

func saveChunk(chunkID string, fileHash string, data []byte) error {
	// Use the file hash to organize chunks.
	chunksDir := fmt.Sprintf("chunks/%s", fileHash)
	err := os.MkdirAll(chunksDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	chunkPath := fmt.Sprintf("%s/%s", chunksDir, chunkID)
	return os.WriteFile(chunkPath, data, 0644)
}
