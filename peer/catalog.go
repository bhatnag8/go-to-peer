// Package peer manages peer connectivity and server file catalog functionality.
package peer

import (
	"fmt"
	"go-to-peer/file"
	"go-to-peer/util"
	"os"
	"path/filepath"
	//"strings"
)

// FileCatalog represents the catalog of files available on the server.
type FileCatalog struct {
	Files []FileMetadata `json:"files"` // List of file metadata.
}

// FileMetadata represents metadata about a file available for sharing.
type FileMetadata struct {
	Name   string   `json:"name"`   // File name.
	Size   int64    `json:"size"`   // File size in bytes.
	Chunks []string `json:"chunks"` // List of chunk IDs for the file.
	Hash   string   `json:"hash"`   // Hash of the entire file for integrity verification.
}

// CreateCatalog generates a catalog from files stored in a given directory.
// It splits the files into chunks and populates the catalog.
//
// Parameters:
// - directory: The directory containing the files to catalog.
//
// Returns:
// - *FileCatalog: A pointer to the generated file catalog.
// - error: An error if catalog creation fails.
// CreateCatalog generates a catalog from files stored in a given directory.
func createCatalog(directory string) (*FileCatalog, error) {
	catalog := &FileCatalog{}
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(directory, entry.Name())
		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil {
			util.Logger.Printf("Failed to stat file %s: %v", filePath, statErr)
			continue
		}

		hash := util.CalculateFileHash(filePath)
		err := file.SplitFile(filePath, hash) // Updated SplitFile to use hash-based directories.
		if err != nil {
			return nil, fmt.Errorf("failed to split file %s: %w", entry.Name(), err)
		}

		chunks := generateChunkList(hash)
		catalog.Files = append(catalog.Files, FileMetadata{
			Name:   entry.Name(),
			Size:   fileInfo.Size(),
			Hash:   hash,
			Chunks: chunks,
		})
	}

	return catalog, nil
}

// generateChunkList creates a list of chunk IDs for a file based on its hash.
func generateChunkList(hash string) []string {
	chunksDir := filepath.Join("chunks", hash)
	chunks := []string{}

	files, err := os.ReadDir(chunksDir)
	if err != nil {
		util.Logger.Printf("Failed to read chunks directory for hash %s: %v", hash, err)
		return chunks
	}

	for _, chunk := range files {
		if !chunk.IsDir() {
			chunks = append(chunks, chunk.Name())
		}
	}
	return chunks
}
