package re

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// EOF represents the end of the file.
const EOF = -1

// parser is a simple regular expression parser.
type parser struct {
	regexp string
	pos    int
	tokens []token
}

// peek returns the next rune and its size in the input string without advancing the position.
// If the end of the string is reached, it returns EOF and -1.
func (p *parser) peek() (rune, int) {
	if p.pos >= len(p.regexp) {
		return EOF, -1
	}

	r, w := utf8.DecodeRuneInString(p.regexp[p.pos:])
	return r, w
}

// next returns the next rune in the input string and advances the position by the rune's width.
// If the end of the string is reached, it returns EOF.
func (p *parser) next() rune {
	r, w := p.peek()
	if r != EOF {
		p.pos += w
	}
	return r
}

// parse processes the entire regular expression string, parsing it into its constituent parts.
// It returns an error if any part of the regular expression is invalid.
func (p *parser) parse() error {
	for p.pos < len(p.regexp) {
		err := p.parseRe()
		if err != nil {
			return err
		}
	}
	return nil
}

// parseRe processes the regular expression string by parsing individual characters.
// It returns an error if any part of the regular expression is invalid or if an unexpected EOF is encountered.
func (p *parser) parseRe() error {
	var err error
	nextRune, _ := p.peek()
	switch nextRune {
	case EOF:
		return nil
	case '\\':
		err = p.parseMetaChar()
	default:
		err = p.parseLiteral()
	}

	if err != nil {
		return err
	}

	return nil
}

// parseLiteral reads the next rune from the input string and appends it as a charToken to the tokens slice.
// If the end of the input string is reached, it returns an error indicating an unexpected EOF.
func (p *parser) parseLiteral() error {
	r := p.next()
	if r == EOF {
		return errors.New("unexpected EOF")
	}

	token := literalToken{char: r}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseMetaChar reads the next rune from the input string and appends it as a meta character token to the tokens slice.
// If the end of the input string is reached, it returns an error indicating an unexpected EOF.
// If the meta character is not supported, it returns an error.
func (p *parser) parseMetaChar() error {
	nextChar := p.next()
	if nextChar == EOF {
		return errors.New("unexpected EOF while parsing meta character")
	}

	var token token
	nextChar = p.next()
	switch nextChar {
	case 'd':
		token = digitToken{}
	default:
		return fmt.Errorf("unsupported meta character: \\%c", nextChar)
	}

	p.tokens = append(p.tokens, token)
	return nil
}

// token represents a regular expression token.
type token interface {
	toNfa() *nfa
}

// literalToken represents a character token.
type literalToken struct {
	char rune
}

// toNfa converts the literal token to an NFA.
func (t literalToken) toNfa() *nfa {
	start := &state{edges: make(map[rune][]*state)}
	end := &state{isFinal: true}
	start.edges[t.char] = []*state{end}
	return &nfa{start, end}
}

// digitToken represents a digit token.
type digitToken struct{}

// toNfa converts the digit token to an NFA.
func (t digitToken) toNfa() *nfa {
	start := &state{edges: make(map[rune][]*state)}
	end := &state{isFinal: true}
	for r := '0'; r <= '9'; r++ {
		start.edges[r] = []*state{end}
	}
	return &nfa{start, end}
}

// state represents a state in the NFA.
type state struct {
	edges   map[rune][]*state
	epsilon []*state
	isFinal bool
}

// nfa represents a Non-deterministic Finite Automaton.
type nfa struct {
	start *state
	end   *state
}

// buildNfa builds an NFA from the parsed regular expression.
func buildNfa(tokens []token) *nfa {
	var nfa *nfa
	for _, token := range tokens {
		nextNfa := token.toNfa()
		if nfa == nil {
			nfa = nextNfa
		} else {
			nfa.end.epsilon = append(nfa.end.epsilon, nextNfa.start)
			nfa.end.isFinal = false
			nfa.end = nextNfa.end
		}
	}
	return nfa
}

// matches takes a string s as input and recursively searches the NFA to determine if it reaches a final state.
// It returns true if the NFA can match the part of input string, otherwise false.
func (n *nfa) matches(s string) bool {
	var checkMatch func(state *state, s string) bool
	checkMatch = func(state *state, s string) bool {
		if state.isFinal || len(s) == 0 {
			return true
		}

		r, w := utf8.DecodeRuneInString(s)
		if st := state.edges[r]; st != nil {
			if checkMatch(st[0], s[w:]) {
				return true
			}
		}

		for _, st := range state.epsilon {
			if checkMatch(st, s) {
				return true
			}
		}

		return false
	}

	return checkMatch(n.start, s)
}

// Match checks if the given line contains any match of the specified regular expression pattern.
// It returns true if a match is found, otherwise false. If the pattern is invalid, it returns an error.
func Match(line, pattern string) (bool, error) {
	if pattern == "" {
		return true, nil
	}

	p := parser{regexp: pattern}
	err := p.parse()
	if err != nil {
		return false, err
	}

	nfa := buildNfa(p.tokens)
	for len(line) > 0 {
		if nfa.matches(line) {
			return true, nil
		}
		_, runeSize := utf8.DecodeRuneInString(line)
		line = line[runeSize:]
	}

	return false, nil
}
