package lisp

import "fmt"

// tokenStream allows one to iterate, seek and peek over a stream of tokens.
//
// Allowing peek and seek over a raw slice of tokens allows a parser to revert its progress through the stream
// NOTE: not very useful perhaps in LISP? Seek and Peekable streams are useful for ambiguous syntax, but we don't have that here at all.
type tokenStream struct {
	tokens []token
	idx    int
}

func newTokenStream(tokens []token) *tokenStream {
	return &tokenStream{
		tokens: tokens,
		idx:    0,
	}
}

func (ts *tokenStream) next() (token, bool) {
	if ts.idx == len(ts.tokens) {
		return token{}, false
	}

	ts.idx++
	return ts.tokens[ts.idx-1], true
}

func (ts *tokenStream) position() int {
	return ts.idx
}

func (ts *tokenStream) seek(idx int) {
	ts.idx = idx
}

func (ts *tokenStream) peek() (token, bool) {
	if ts.idx == len(ts.tokens) {
		return token{}, false
	}
	return ts.tokens[ts.idx], true
}

func (ts *tokenStream) close() error {
	if ts.idx != len(ts.tokens) {
		return fmt.Errorf("trailing tokens")
	}
	return nil
}

func (ts *tokenStream) nextTokenIs(typ tokenType, val string) bool {
	switch {
	case ts.eof():
		return false
	case ts.tokens[ts.idx].typ != typ:
		return false
	case ts.tokens[ts.idx].val != val:
		return false
	default:
		return true
	}
}

func (ts *tokenStream) nextTokenTypeIs(typ tokenType) bool {
	switch {
	case ts.eof():
		return false
	case ts.tokens[ts.idx].typ != typ:
		return false
	default:
		return true
	}
}

func (ts *tokenStream) eof() bool {
	return len(ts.tokens) <= ts.idx
}
