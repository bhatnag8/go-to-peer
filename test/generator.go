// Package testfiles provides functionality to generate random test files
// for simulating various file sizes in the application.
package test

import (
	"fmt"       // Formatted I/O library
	"math/rand" // For generating random data
	"os"        // OS-level file handling
	"time"      // Used for seeding random data
)

// Initialize the random seed for generating random data.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// CreateTestDataDir ensures the "testdata" directory exists for storing generated files.
func CreateTestDataDir() error {
	if _, err := os.Stat("testdata"); os.IsNotExist(err) {
		err := os.Mkdir("testdata", 0755)
		if err != nil {
			return fmt.Errorf("failed to create testdata directory: %w", err)
		}
	}
	return nil
}

// DeleteTestDataDir removes the "testdata" directory and its contents.
func DeleteTestDataDir() error {
	if err := os.RemoveAll("testdata"); err != nil {
		return fmt.Errorf("failed to delete testdata directory: %w", err)
	}
	return nil
}

// GenerateRandomFile creates a random file of the specified size (in bytes)
// inside the "testdata" directory with a random name.
func GenerateRandomFile(size int64) (string, error) {
	// Ensure the "testdata" directory exists.
	err := CreateTestDataDir()
	if err != nil {
		return "", err
	}

	// Create a file with a random name.
	fileName := fmt.Sprintf("testdata/random_%d.dat", rand.Int63())
	file, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file %s: %v\n", fileName, closeErr)
		}
	}()

	// Generate random data and write it to the file.
	data := make([]byte, 4096) // 4 KB buffer
	var bytesWritten int64
	for bytesWritten < size {
		remaining := size - bytesWritten
		toWrite := int64(len(data))
		if remaining < toWrite {
			toWrite = remaining
		}
		_, err := rand.Read(data[:toWrite])
		if err != nil {
			return "", fmt.Errorf("failed to generate random data: %w", err)
		}
		_, writeErr := file.Write(data[:toWrite])
		if writeErr != nil {
			return "", fmt.Errorf("failed to write random data to file: %w", writeErr)
		}
		bytesWritten += toWrite
	}

	return fileName, nil
}
