package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateHash generates a SHA-256 hash for the given data.
func CalculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// CalculateFileHash computes the SHA-256 hash of the given file.
//
// Parameters:
// - filePath: The path to the file.
//
// Returns:
// - string: The computed hash as a hexadecimal string.
// - error: An error object if any issues occur while accessing the file.
func CalculateFileHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hasher := sha256.New()
	_, _ = io.Copy(hasher, file)
	return hex.EncodeToString(hasher.Sum(nil))
}
