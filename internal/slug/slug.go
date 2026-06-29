package slug

import (
	"crypto/rand"
	"strings"
)

const defaultLength = 7

const base62Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func Generate() string {
	bytes := make([]byte, defaultLength)
	_, _ = rand.Read(bytes)

	var sb strings.Builder
	sb.Grow(defaultLength)
	for _, b := range bytes {
		sb.WriteByte(base62Alphabet[int(b)%len(base62Alphabet)])
	}

	return sb.String()
}

func IsValid(s string) bool {
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
