package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken creates a secure random string for verification/resets
func GenerateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
