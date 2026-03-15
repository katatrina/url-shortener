package shortcode

import "testing"

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
		name   string
		length int
	}{
		{"length 5", 5},
		{"length 7", 7},
		{"length 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GenerateWithLength(tt.length)

			if len(code) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(code))
			}
		})
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	// Generate 1000 codes and verify no duplicates.
	// With 62^7 ≈ 3.5 trillion combinations, collision in 1000 codes
	// would indicate a broken random source.
	seen := make(map[string]bool)
	count := 1000

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
		name  string
		code  string
		valid bool
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
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.code)
			if got != tt.valid {
				t.Errorf("IsValid(%q) = %v, want %v", tt.code, got, tt.valid)
			}
		})
	}
}
