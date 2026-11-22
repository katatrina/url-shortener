package request

import (
	"strings"
	"testing"
)

func TestASCIIEmail(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"a@b.co", true},
		{"first.last+tag@example.com", true},
		{"user_name-1@sub.example.vn", true},
		{"a..b@c.com", true}, // allowed by the WHATWG grammar
		{"!#$%&'*+/=?^_`{|}~-@example.com", true},
		{"a@x" + strings.Repeat("a", 61) + "x.com", true}, // 63-char label

		{"a@b", false},            // no dot in the domain
		{"user@localhost", false}, // no dot in the domain
		{"a@b.", false},           // empty last label
		{"a@.b.com", false},       // empty label
		{"a@-b.com", false},       // label starts with '-'
		{"a@b-.com", false},       // label ends with '-'
		{"a@b_c.com", false},      // '_' not allowed in the domain
		{"a b@c.com", false},
		{`"a b"@example.com`, false}, // quoted local part
		{"tên@ví-dụ.vn", false},      // non-ASCII
		{"a@b@c.com", false},
		{"@example.com", false},
		{"a@x" + strings.Repeat("a", 62) + "x.com", false}, // 64-char label
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := validate.Var(tt.in, "ascii_email") == nil
			if got != tt.want {
				t.Errorf("ascii_email(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
