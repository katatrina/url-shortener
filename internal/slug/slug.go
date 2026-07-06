package slug

import (
	"crypto/rand"
	"strings"
)

const defaultLength = 7

const base62Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Generate returns a random base62 slug of defaultLength characters.
func Generate() string {
	bytes := make([]byte, defaultLength)
	// crypto/rand.Read never returns an error (Go 1.24+).
	_, _ = rand.Read(bytes)

	var sb strings.Builder
	sb.Grow(defaultLength)
	for _, b := range bytes {
		sb.WriteByte(base62Alphabet[int(b)%len(base62Alphabet)])
	}

	return sb.String()
}

// IsValid reports whether s is a valid slug: a non-empty ASCII string of [A-Za-z0-9-_].
func IsValid(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if !isAllowedByte(s[i]) {
			return false
		}
	}

	return true
}

func isAllowedByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	case b == '-' || b == '_':
		return true
	default:
		return false
	}
}
