package shortcode

import (
	"crypto/rand"
	"strings"
)

// alphabet contains 62 characters: a-z, A-Z, 0-9.
// Using base62 gives us 62^7 ≈ 3.5 trillion possible combinations for a 7-char code.
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// DefaultLength is the default length of generated short codes.
// 7 characters give 62^7 ≈ 3.5 trillion combinations — collision probability
// is negligible until billions of URLs are stored.
const DefaultLength = 7

// Generate creates a cryptographically random short code.
//
// Why crypto/rand instead of math/rand?
// math/rand is predictable if you know the seed — an attacker could guess
// the next short codes. crypto/rand reads from /dev/urandom (on Linux),
// which is truly unpredictable.
func Generate() string {
	return GenerateWithLength(DefaultLength)
}

// GenerateWithLength creates a random short code with the specified length.
func GenerateWithLength(length int) string {
	// Read random bytes from the OS.
	bytes := make([]byte, length)
	// This will never return an error, and crash program if failed.
	_, _ = rand.Read(bytes)

	// Map each byte to a character in our alphabet.
	// Using modulo introduces a tiny bias (256 % 62 != 0), but for URL
	// shortening this is perfectly acceptable. If you needed cryptographic
	// uniformity, you'd use rejection sampling instead.
	var sb strings.Builder
	sb.Grow(length)
	for _, b := range bytes {
		sb.WriteByte(alphabet[int(b)%len(alphabet)])
	}

	return sb.String()
}

// IsValid checks if a string is a valid short code (only base62 characters).
// Useful for validating custom aliases provided by users.
func IsValid(code string) bool {
	if len(code) == 0 {
		return false
	}

	for _, c := range code {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}

	return true
}
