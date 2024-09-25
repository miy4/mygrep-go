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
		{"a", "[^a-c]", false, nil, false},
		{"b", "[^a-c]", false, nil, false},
		{"d", "[^a-c]", true, nil, false},
		{"a", "[^c-a]", false, errors.New("invalid range: c-a"), true},
		{"a", "[^a-]", false, nil, false},
		{"b", "[^a-]", true, nil, false},
		{"-", "[^a-]", false, nil, false},
		{"dog", "[^abc]", true, nil, false},
		{"cab", "[^abc]", false, nil, false},
		{"1 apple", "\\d apple", true, nil, false},
		{"1 orange", "\\d apple", false, nil, false},
		{"100 apple", "\\d\\d\\d apple", true, nil, false},
		{"1 apple", "\\d\\d\\d apple", false, nil, false},
		{"3 dogs", "\\d \\w\\w\\ws", true, nil, false},
		{"4 cats", "\\d \\w\\w\\ws", true, nil, false},
		{"1 dog", "\\d \\w\\w\\ws", false, nil, false},
		{"sally has 3 apples", "\\d apple", true, nil, false},
		{"sally has 1 orange", "\\d apple", false, nil, false},
		{"sally has 12 apples", "\\d \\\\d\\\\d apples", false, nil, false},
		{"log file", "^log", true, nil, false},
		{"error log", "^log", false, nil, false},
		{"dog", "dog$", true, nil, false},
		{"dogs", "dog$", false, nil, false},
		{"eels", "e+", true, nil, false},
		{"els", "e+", true, nil, false},
		{"ls", "e+", false, nil, false},
		{"dogs", "dogs?", true, nil, false},
		{"dog", "dogs?", true, nil, false},
		{"cat", "dogs?", false, nil, false},
		{"dog", "d.g", true, nil, false},
		{"dog", "c.g", false, nil, false},
		{"cat", "c.t", true, nil, false},
		{"cot", "c.t", true, nil, false},
		{"car", "c.t", false, nil, false},
		{"cat", "(cat|dog)", true, nil, false},
		{"dog", "(cat|dog)", true, nil, false},
		{"apple", "(cat|dog)", false, nil, false},
		{"a cat", "a (cat|dog)", true, nil, false},
		{"a dog", "a (cat|dog)", true, nil, false},
		{"a cow", "a (cat|dog)", false, nil, false},
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
