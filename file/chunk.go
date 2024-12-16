// Package file contains utilities for file chunking and reconstruction.
package file

// Import statements:
// - "os": For file I/O operations (e.g., reading and writing files).
// - "io": For general file stream handling.
// - "fmt": For formatted error messages.
import (
	"encoding/json"
	"fmt" // Formatted I/O library
	"io"  // Input/Output utility library
	"os"  // OS-level file handling functions
	"path/filepath"
	//"strings"
)

// ChunkSize defines the size of each file chunk in bytes (1MB).
const ChunkSize = 1 * 1024 * 1024

// SplitFile splits a given file into chunks of fixed size.
// The chunks are stored in the "chunks" directory with a numbered naming scheme.
// SplitFile splits a given file into chunks of fixed size.

// FileMetadata represents metadata for a file.
type FileMetadata struct {
	Name   string   `json:"name"`   // Original file name
	Size   int64    `json:"size"`   // File size in bytes
	Chunks []string `json:"chunks"` // List of chunk IDs
	Hash   string   `json:"hash"`   // File hash
}

func SplitFile(filePath string, fileHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	chunksDir := filepath.Join("chunks", fileHash)
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	var chunkIDs []string
	buffer := make([]byte, ChunkSize)
	chunkIndex := 0
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}
		if bytesRead == 0 {
			break
		}

		chunkID := fmt.Sprintf("chunk_%d", chunkIndex)
		chunkPath := filepath.Join(chunksDir, chunkID)
		chunkFile, err := os.Create(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to create chunk file: %w", err)
		}

		if _, err := chunkFile.Write(buffer[:bytesRead]); err != nil {
			return fmt.Errorf("error writing chunk file: %w", err)
		}
		chunkFile.Close()

		chunkIDs = append(chunkIDs, chunkID)
		chunkIndex++
	}

	// Create metadata.json
	metadata := FileMetadata{
		Name:   fileInfo.Name(),
		Size:   fileInfo.Size(),
		Chunks: chunkIDs,
		Hash:   fileHash,
	}

	metadataPath := filepath.Join(chunksDir, "metadata.json")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()

	if err := json.NewEncoder(metadataFile).Encode(metadata); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// ReconstructFile reconstructs the original file from its chunks.
// It reads chunks from the "chunks" directory and combines them into a single output file.
// ReconstructFile reconstructs the original file from its chunks.
// It reads chunks from the "chunks/<file_name>" directory and combines them into a single output file.
func ReconstructFile(outputDir string, fileHash string) error {
	chunksDir := filepath.Join("chunks", fileHash)
	metadataFilePath := filepath.Join(chunksDir, "metadata.json")

	// Open and parse the metadata file.
	metadataFile, err := os.Open(metadataFilePath)
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer metadataFile.Close()

	var metadata FileMetadata
	if err := json.NewDecoder(metadataFile).Decode(&metadata); err != nil {
		return fmt.Errorf("failed to parse metadata file: %w", err)
	}

	// Use the original file name from metadata.
	originalName := metadata.Name
	if originalName == "" {
		return fmt.Errorf("original file name is missing in metadata")
	}

	// Set the output file path.
	outputFilePath := filepath.Join(outputDir, originalName)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Reconstruct the file from chunks.
	for _, chunkID := range metadata.Chunks {
		chunkPath := filepath.Join(chunksDir, chunkID)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to open chunk file %s: %w", chunkPath, err)
		}

		// Write the chunk data to the output file.
		if _, err := io.Copy(outputFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("failed to write chunk data to output file: %w", err)
		}
		chunkFile.Close()
	}

	return nil
}
