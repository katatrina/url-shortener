package shortcode

import (
	"crypto/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// DefaultLength is the default length of generated short codes.
const DefaultLength = 7

// Generate creates a cryptographically random short code.
func Generate() string {
	return GenerateWithLength(DefaultLength)
}

// GenerateWithLength creates a random short code with the specified length.
func GenerateWithLength(length int) string {
	if length <= 0 {
		length = DefaultLength
	}

	bytes := make([]byte, length)
	rand.Read(bytes)

	var sb strings.Builder
	sb.Grow(length)
	for _, b := range bytes {
		sb.WriteByte(alphabet[int(b)%len(alphabet)])
	}

	return sb.String()
}

// IsValid check if a string is a valid short code (only base62 characters).
func IsValid(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}

	return true
}
