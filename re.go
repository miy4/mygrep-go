package re

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const EOF = -1     // End of file
const BOS = '\x02' // Beginning of string
const EOS = '\x03' // End of string

// parser is a simple regular expression parser.
type parser struct {
	regexp string
	pos    int
	tokens []token
	done   bool
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
	for p.pos < len(p.regexp) && !p.done {
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
	nextRune, runeSize := p.peek()
	switch nextRune {
	case EOF:
		return nil
	case '(':
		err = p.parseGroup()
	case ')':
		err = p.parseClosingGroup()
	case '|':
		err = p.parseOr()
	case '[':
		nextNextRune, _ := utf8.DecodeRuneInString(p.regexp[p.pos+runeSize:])
		if nextNextRune == '^' {
			err = p.parseNegativeSet()
		} else {
			err = p.parsePositiveSet()
		}
	case '+':
		err = p.parsePlus()
	case '*':
		err = p.parseStar()
	case '?':
		err = p.parseOptional()
	case '.':
		err = p.parseWildcard()
	case '^':
		err = p.parseBeginningOfString()
	case '$':
		err = p.parseEndOfString()
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
	case 'w':
		token = wordToken{}
	case '\\':
		token = literalToken{char: '\\'}
	default:
		return fmt.Errorf("unsupported meta character: \\%c", nextChar)
	}

	p.tokens = append(p.tokens, token)
	return nil
}

// parsePositiveSet parses a positive set from the input string. It expects the input to start with '[' and contain a closing ']'.
// It reads runes from the input string and appends them to the setItems slice. If a range (e.g., 'a-z') is detected, it handles it appropriately.
// If the input string ends unexpectedly or if the set is not properly closed, it returns an error.
func (p *parser) parsePositiveSet() error {
	if p.next() != '[' {
		return errors.New("expected '[' at the beginning of positive set")
	} else if !strings.ContainsRune(p.regexp[p.pos:], ']') {
		return errors.New("unclosed '[' in positive set")
	}

	var previousChar rune
	setItems := make([]rune, 0)
	for currentChar := p.next(); currentChar != ']'; currentChar = p.next() {
		if currentChar == EOF {
			return errors.New("unexpected EOF while parsing positive set")
		}

		if currentChar == '-' && previousChar != 0 {
			rangeStart := previousChar
			rangeEnd := p.next()
			if rangeEnd == EOF {
				return errors.New("unexpected EOF while parsing range in positive set")
			} else if rangeEnd == ']' {
				setItems = append(setItems, previousChar, '-')
				break
			}

			if rangeStart > rangeEnd {
				return fmt.Errorf("invalid range: %c-%c", rangeStart, rangeEnd)
			}

			for ch := rangeStart; ch <= rangeEnd; ch++ {
				setItems = append(setItems, ch)
			}
			previousChar = 0
		} else {
			setItems = append(setItems, currentChar)
			previousChar = currentChar
		}
	}

	if len(setItems) == 0 {
		return errors.New("empty positive set")
	}
	p.tokens = append(p.tokens, positiveSetToken{setItems})
	return nil
}

// parseNegativeSet parses a negative set from the input string. It expects the input to start with '[^' and contain a closing ']'.
// It reads runes from the input string and appends them to the setItems slice. If a range (e.g., 'a-z') is detected, it handles it appropriately.
// If the input string ends unexpectedly or if the set is not properly closed, it returns an error.
func (p *parser) parseNegativeSet() error {
	if p.next() != '[' || p.next() != '^' {
		return errors.New("expected '[^' at the beginning of negative set")
	} else if !strings.ContainsRune(p.regexp[p.pos:], ']') {
		return errors.New("unclosed '[' in negative set")
	}

	var previousChar rune
	setItems := make([]rune, 0)
	for currentChar := p.next(); currentChar != ']'; currentChar = p.next() {
		if currentChar == EOF {
			return errors.New("unexpected EOF while parsing negative set")
		}

		if currentChar == '-' && previousChar != 0 {
			rangeStart := previousChar
			rangeEnd := p.next()
			if rangeEnd == EOF {
				return errors.New("unexpected EOF while parsing range in negative set")
			} else if rangeEnd == ']' {
				setItems = append(setItems, previousChar, '-')
				break
			}

			if rangeStart > rangeEnd {
				return fmt.Errorf("invalid range: %c-%c", rangeStart, rangeEnd)
			}

			for ch := rangeStart; ch <= rangeEnd; ch++ {
				setItems = append(setItems, ch)
			}

			previousChar = 0
		} else {
			setItems = append(setItems, currentChar)
			previousChar = currentChar
		}
	}

	if len(setItems) == 0 {
		return errors.New("empty negative set")
	}

	p.tokens = append(p.tokens, negativeSetToken{setItems})
	return nil
}

// parseBeginningOfString parses the beginning of string token '^' from the input string.
func (p *parser) parseBeginningOfString() error {
	if p.next() != '^' {
		return errors.New("expected '^' at the beginning of string")
	}

	token := beginningOfStringToken{}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseEndOfString parses the end of string token '$' from the input string.
func (p *parser) parseEndOfString() error {
	if p.next() != '$' {
		return errors.New("expected '$' at the end of string")
	}

	token := endOfStringToken{}
	p.tokens = append(p.tokens, token)
	return nil
}

// parsePlus parses the one or more quantifier '+' from the input string.
func (p *parser) parsePlus() error {
	if p.next() != '+' {
		return errors.New("expected '+' after character")
	} else if len(p.tokens) == 0 {
		return errors.New("no character to apply '+' to")
	}

	lastToken := p.tokens[len(p.tokens)-1]
	p.tokens = p.tokens[:len(p.tokens)-1]
	token := plusToken{payload: lastToken}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseStar parses the zero or more quantifier '*' from the input string.
func (p *parser) parseStar() error {
	if p.next() != '*' {
		return errors.New("expected '*' after character")
	} else if len(p.tokens) == 0 {
		return errors.New("no character to apply '*' to")
	}

	lastToken := p.tokens[len(p.tokens)-1]
	p.tokens = p.tokens[:len(p.tokens)-1]
	token := starToken{payload: lastToken}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseOptional parses the zero or one quantifier '?' from the input string.
func (p *parser) parseOptional() error {
	if p.next() != '?' {
		return errors.New("expected '?' after character")
	} else if len(p.tokens) == 0 {
		return errors.New("no character to apply '?' to")
	}

	lastToken := p.tokens[len(p.tokens)-1]
	p.tokens = p.tokens[:len(p.tokens)-1]
	token := optionalToken{payload: lastToken}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseWildcard parses the wildcard '.' from the input string.
func (p *parser) parseWildcard() error {
	if p.next() != '.' {
		return errors.New("expected '.'")
	}

	token := wildcardToken{}
	p.tokens = append(p.tokens, token)
	return nil
}

// parseGroup parses a group of tokens enclosed in parentheses from the input string.
// It expects the input to start with '(' and will return an error if it does not.
// The method creates a new parser instance to parse the group and appends the parsed tokens to the current parser's token list.
func (p *parser) parseGroup() error {
	if p.next() != '(' {
		return errors.New("expected '(' at the beginning of group")
	}

	groupParser := parser{
		regexp: p.regexp,
		pos:    p.pos,
		tokens: []token{groupToken{payload: [][]token{}}},
	}

	err := groupParser.parse()
	if err != nil {
		return err
	}

	p.pos = groupParser.pos
	p.tokens = append(p.tokens, groupParser.tokens...)
	return nil
}

// parseOr parses the '|' character from the input string.
// It expects the input to contain a '|' character and a preceding group of tokens.
// If the '|' character is not found or if there is no preceding group, it returns an error.
// The method then appends the tokens following the '|' to the payload of the preceding group token.
func (p *parser) parseOr() error {
	if p.next() != '|' {
		return errors.New("expected '|'")
	} else if _, ok := p.tokens[0].(groupToken); !ok {
		return errors.New("expected group before '|'")
	}

	previousTokens := p.tokens[1:]
	groupToken := p.tokens[0].(groupToken)
	groupToken.payload = append(groupToken.payload, previousTokens)
	p.tokens = []token{groupToken}
	return nil
}

// parseClosingGroup parses the closing ')' character from the input string.
func (p *parser) parseClosingGroup() error {
	if p.next() != ')' {
		return errors.New("expected ')'")
	} else if _, ok := p.tokens[0].(groupToken); !ok {
		return errors.New("expected group before ')'")
	}

	previousTokens := p.tokens[1:]
	groupToken := p.tokens[0].(groupToken)
	groupToken.payload = append(groupToken.payload, previousTokens)
	p.tokens = []token{groupToken}
	p.done = true
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

// wordToken represents an alphanumeric character token.
type wordToken struct{}

// toNfa converts the word token to an NFA.
func (t wordToken) toNfa() *nfa {
	start := &state{edges: make(map[rune][]*state)}
	end := &state{isFinal: true}
	for r := 'a'; r <= 'z'; r++ {
		start.edges[r] = []*state{end}
	}
	for r := 'A'; r <= 'Z'; r++ {
		start.edges[r] = []*state{end}
	}
	for r := '0'; r <= '9'; r++ {
		start.edges[r] = []*state{end}
	}
	start.edges['_'] = []*state{end}
	return &nfa{start, end}
}

// positiveSetToken represents a positive character set token.
type positiveSetToken struct {
	setItems []rune
}

// toNfa converts the positive set token to an NFA.
func (t positiveSetToken) toNfa() *nfa {
	start := &state{edges: make(map[rune][]*state)}
	end := &state{isFinal: true}
	for _, r := range t.setItems {
		start.edges[r] = []*state{end}
	}
	return &nfa{start, end}
}

// negativeSetToken represents a negative character set token.
type negativeSetToken struct {
	setItems []rune
}

// toNfa converts the negative set token to an NFA.
func (t negativeSetToken) toNfa() *nfa {
	start := &state{edges: make(map[rune][]*state)}
	end := &state{isFinal: true}
	deadEnd := &state{}
	for _, r := range t.setItems {
		start.edges[r] = []*state{deadEnd}
	}
	start.anyChar = []*state{end}
	return &nfa{start, end}
}

// beginningOfStringToken represents the beginning of string token.
type beginningOfStringToken struct{}

// toNfa converts the beginning of string token to an NFA.
func (t beginningOfStringToken) toNfa() *nfa {
	start := &state{control: make(map[rune][]*state)}
	end := &state{isFinal: true}
	start.control[BOS] = []*state{end}
	return &nfa{start, end}
}

// endOfStringToken represents the end of string token.
type endOfStringToken struct{}

// toNfa converts the end of string token to an NFA.
func (t endOfStringToken) toNfa() *nfa {
	start := &state{control: make(map[rune][]*state)}
	end := &state{isFinal: true}
	start.control[EOS] = []*state{end}
	return &nfa{start, end}
}

// plusToken represents an one or more quantifier token.
type plusToken struct {
	payload token
}

// toNfa converts the plus token to an NFA.
func (t plusToken) toNfa() *nfa {
	nfa := t.payload.toNfa()
	nfa.end.epsilon = append(nfa.end.epsilon, nfa.start)
	return nfa
}

// starToken represents a zero or more quantifier token.
type starToken struct {
	payload token
}

// toNfa converts the star token to an NFA.
func (t starToken) toNfa() *nfa {
	nfa := t.payload.toNfa()
	nfa.end.epsilon = append(nfa.end.epsilon, nfa.start)
	nfa.start.epsilon = append(nfa.start.epsilon, nfa.end)
	return nfa
}

// optionalToken represents a zero or one quantifier token.
type optionalToken struct {
	payload token
}

// toNfa converts the optional token to an NFA.
func (t optionalToken) toNfa() *nfa {
	nfa := t.payload.toNfa()
	nfa.start.epsilon = append(nfa.start.epsilon, nfa.end)
	return nfa
}

// wildcardToken represents a wildcard token.
type wildcardToken struct{}

// toNfa converts the wildcard token to an NFA.
func (t wildcardToken) toNfa() *nfa {
	start := &state{}
	end := &state{isFinal: true}
	start.anyChar = []*state{end}
	return &nfa{start, end}
}

// groupToken represents a group of tokens.
type groupToken struct {
	payload [][]token
}

// toNfa converts the group token to an NFA.
func (t groupToken) toNfa() *nfa {
	start := &state{epsilon: []*state{}}
	end := &state{isFinal: true}
	for _, tokens := range t.payload {
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

		start.epsilon = append(start.epsilon, nfa.start)
		nfa.end.epsilon = append(nfa.end.epsilon, end)
		nfa.end.isFinal = false
	}

	return &nfa{start, end}
}

// state represents a state in the NFA.
type state struct {
	edges   map[rune][]*state
	control map[rune][]*state
	anyChar []*state
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
		if state.isFinal {
			return true
		}

		r, w := utf8.DecodeRuneInString(s)
		if unicode.IsPrint(r) {
			if st := state.edges[r]; st != nil {
				if checkMatch(st[0], s[w:]) {
					return true
				}
			} else if state.anyChar != nil {
				if checkMatch(state.anyChar[0], s[w:]) {
					return true
				}
			}
		} else {
			if st := state.control[r]; st != nil {
				if checkMatch(st[0], s[w:]) {
					return true
				}
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

// stringSource prepares the input string for matching by replacing newline characters with the beginning-of-string character.
// It also prepends the BOS character to the start of the string.
func stringSource(input string) string {
	preparedString := strings.ReplaceAll(input, "\n", string(EOS)+string(BOS))
	preparedString = string(BOS) + preparedString + string(EOS)
	return preparedString
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

	line = stringSource(line)
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
