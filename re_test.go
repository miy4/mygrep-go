package re

import (
	"errors"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		line          string
		pattern       string
		expected      bool
		err           error
		errorExpected bool
	}{
		{"a", "a", true, nil, false},
		{"b", "a", false, nil, false},
		{"", "a", false, nil, false},
		{"a", "", true, nil, false},
		{"3", "d", false, nil, false},
		{"3", "\\d", true, nil, false},
		{"d", "\\d", false, nil, false},
		{"apple123", "\\d", true, nil, false},
		{"altern8", "altern\\d", true, nil, false},
		{"a", "\\w", true, nil, false},
		{"Z", "\\w", true, nil, false},
		{"0", "\\w", true, nil, false},
		{"9", "\\w", true, nil, false},
		{"_", "\\w", true, nil, false},
		{"foo101", "\\w", true, nil, false},
		{"$!?", "\\w", false, nil, false},
		{"a", "\\@", false, errors.New("unsupported meta character: \\@"), true},
		{"apple", "[abc]", true, nil, false},
		{"dog", "[abc]", false, nil, false},
		{"a", "[a-c]", true, nil, false},
		{"b", "[a-c]", true, nil, false},
		{"d", "[a-c]", false, nil, false},
		{"a", "[c-a]", false, errors.New("invalid range: c-a"), true},
		{"a", "[a-]", true, nil, false},
		{"b", "[a-]", false, nil, false},
		{"-", "[a-]", true, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.line+"_"+tt.pattern, func(t *testing.T) {
			result, err := Match(tt.line, tt.pattern)
			if result != tt.expected {
				t.Errorf("Match(%q, %q) = %v, %v; want %v, %v", tt.line, tt.pattern, result, err, tt.expected, tt.err)
			}

			if tt.errorExpected && (err == nil || err.Error() != tt.err.Error()) {
				t.Errorf("Match(%q, %q) = %v, %v; want %v, %v", tt.line, tt.pattern, result, err, tt.expected, tt.err)
			} else if !tt.errorExpected && err != nil {
				t.Errorf("Match(%q, %q) = %v, %v; want %v, %v", tt.line, tt.pattern, result, err, tt.expected, tt.err)
			}
		})
	}
}
