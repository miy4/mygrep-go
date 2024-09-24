package re

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		line     string
		pattern  string
		expected bool
		err      error
	}{
		{"a", "a", true, nil},
		{"b", "a", false, nil},
		{"", "a", false, nil},
		{"a", "", true, nil},
		{"3", "d", false, nil},
		{"3", "\\d", true, nil},
		{"d", "\\d", false, nil},
		{"apple123", "\\d", true, nil},
		{"altern8", "altern\\d", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.line+"_"+tt.pattern, func(t *testing.T) {
			result, err := Match(tt.line, tt.pattern)
			if result != tt.expected || err != tt.err {
				t.Errorf("Match(%q, %q) = %v, %v; want %v, %v", tt.line, tt.pattern, result, err, tt.expected, tt.err)
			}
		})
	}
}
