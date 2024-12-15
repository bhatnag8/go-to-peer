// Package peer manages peer connectivity and server file catalog functionality.
package peer

import (
	"fmt"
	"go-to-peer/file"
	"os"
	"path/filepath"
	"strings"
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
func createCatalog(directory string) (*FileCatalog, error) {
	catalog := &FileCatalog{}

	// Walk through the directory and process each file.
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}
		if info.IsDir() {
			return nil // Skip directories.
		}

		// Generate chunks for the file.
		err = file.SplitFile(path)
		if err != nil {
			return fmt.Errorf("error splitting file %s: %w", path, err)
		}

		// Create chunk IDs for the file.
		chunks := []string{}
		fileName := filepath.Base(path)
		fileChunkDir := filepath.Join("chunks", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		chunkFiles, err := filepath.Glob(filepath.Join(fileChunkDir, "chunk_*"))
		if err != nil {
			return fmt.Errorf("error retrieving chunks for file %s: %w", fileName, err)
		}
		for _, chunk := range chunkFiles {
			chunks = append(chunks, filepath.Base(chunk))
		}

		// Add file metadata to the catalog.
		catalog.Files = append(catalog.Files, FileMetadata{
			Name:   fileName,
			Size:   info.Size(),
			Chunks: chunks,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return catalog, nil
}
