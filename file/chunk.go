// Package file contains utilities for file chunking and reconstruction.
package file

// Import statements:
// - "os": For file I/O operations (e.g., reading and writing files).
// - "io": For general file stream handling.
// - "fmt": For formatted error messages.
import (
	"fmt" // Formatted I/O library
	"io"  // Input/Output utility library
	"os"  // OS-level file handling functions
)

// ChunkSize defines the size of each file chunk in bytes (1MB).
const ChunkSize = 1 * 1024 * 1024

// SplitFile splits a given file into chunks of fixed size.
// The chunks are stored in the "chunks" directory with a numbered naming scheme.
func SplitFile(filePath string) error {
	// Validate file path input.
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Open the source file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close file: %v\n", err)
		}
	}()

	// Create or clear the directory for storing chunks.
	if _, err := os.Stat("chunks"); err == nil {
		if err := os.RemoveAll("chunks"); err != nil {
			return fmt.Errorf("failed to clear existing chunks: %w", err)
		}
	}
	if err := os.Mkdir("chunks", 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Read the file and write chunks.
	buffer := make([]byte, ChunkSize)
	chunkIndex := 0

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}
		if bytesRead == 0 {
			break // End of file reached.
		}

		// Create a new chunk file.
		chunkFilePath := fmt.Sprintf("chunks/chunk_%d", chunkIndex)
		chunkFile, err := os.Create(chunkFilePath)
		if err != nil {
			return fmt.Errorf("failed to create chunk file: %w", err)
		}

		// Write the chunk data to the file.
		_, writeErr := chunkFile.Write(buffer[:bytesRead])
		if closeErr := chunkFile.Close(); closeErr != nil {
			return fmt.Errorf("failed to close chunk file: %w", closeErr)
		}
		if writeErr != nil {
			return fmt.Errorf("error writing to chunk file: %w", writeErr)
		}

		chunkIndex++ // Increment the chunk index.
	}

	return nil
}

// ReconstructFile reconstructs the original file from its chunks.
// It reads chunks from the "chunks" directory and combines them into a single output file.
func ReconstructFile(outputFilePath string) error {
	// Create the output file for writing.
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Printf("Warning: failed to close output file: %v\n", err)
		}
	}()

	// Read chunks and write them sequentially to the output file.
	chunkIndex := 0
	for {
		chunkFilePath := fmt.Sprintf("chunks/chunk_%d", chunkIndex)
		chunkFile, err := os.Open(chunkFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				break // No more chunks to process.
			}
			return fmt.Errorf("error opening chunk file: %w", err)
		}

		// Copy chunk data to the output file.
		_, copyErr := io.Copy(outputFile, chunkFile)
		if closeErr := chunkFile.Close(); closeErr != nil {
			return fmt.Errorf("failed to close chunk file: %w", closeErr)
		}
		if copyErr != nil {
			return fmt.Errorf("error writing chunk data: %w", copyErr)
		}

		chunkIndex++ // Increment the chunk index.
	}

	return nil
}
