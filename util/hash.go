package util

import (
	"crypto/sha256"
	"fmt"
)

// CalculateHash generates a SHA-256 hash for the given data.
func CalculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
