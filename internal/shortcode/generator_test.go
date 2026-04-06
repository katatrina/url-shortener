package shortcode

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	code := Generate()

	if len(code) != DefaultLength {
		t.Errorf("expected length %d, got %d", DefaultLength, len(code))
	}

	if !IsValid(code) {
		t.Errorf("generated code %q contains invalid characters", code)
	}
}

func TestGenerateWithLength(t *testing.T) {
	tests := []struct {
		name           string
		inputLength    int
		expectedLength int
	}{
		{"length -1", -1, 7},
		{"length 0", 0, 7},
		{"length 5", 5, 5},
		{"length 7", 7, 7},
		{"length 10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GenerateWithLength(tt.inputLength)

			if len(code) != tt.expectedLength {
				t.Errorf("expected length %d, got %d", tt.expectedLength, len(code))
			}
		})
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 1000000

	for range count {
		code := Generate()

		if seen[code] {
			t.Fatalf("duplicate code generated: %s", code)
		}

		seen[code] = true
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid lowercase", "abcdefg", true},
		{"valid uppercase", "ABCDEFG", true},
		{"valid digits", "1234567", true},
		{"valid mixed", "aB3kX9m", true},
		{"empty string", "", false},
		{"contains dash", "abc-def", false},
		{"contains underscore", "abc_def", false},
		{"contains space", "abc def", false},
		{"contains special char", "abc@def", false},
	}

	for _, tt := range tests {
		got := IsValid(tt.code)

		if got != tt.expected {
			t.Errorf("IsValid(%q) = %v, expected %v", tt.code, got, tt.expected)
		}
	}
}
