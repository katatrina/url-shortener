package slug

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	s := Generate()

	if len(s) != defaultLength {
		t.Errorf("Generate() length = %d, want %d", len(s), defaultLength)
	}

	for i := 0; i < len(s); i++ {
		if !strings.ContainsRune(base62Alphabet, rune(s[i])) {
			t.Errorf("Generate() = %q contains non-base62 byte %q", s, s[i])
		}
	}

	if !IsValid(s) {
		t.Errorf("Generate() = %q is not a valid slug", s)
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"mixed allowed chars", "abc-DEF_123", true},
		{"empty", "", false},
		{"space", "abc def", false},
		{"slash", "abc/def", false},
		{"unicode", "café", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.in); got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
