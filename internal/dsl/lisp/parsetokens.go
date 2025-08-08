package lisp

import (
	"fmt"
	"strconv"
)

// ParseTokenStream takes a token slice and converts it into an AST. This AST can then be interpreted into an actual model.
func ParseTokenStream(tokens []token) ([]node, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("token stream is empty")
	}

	ns := []node{}
	ts := newTokenStream(tokens)
	for {
		if _, ok := ts.peek(); !ok {
			break
		}

		n, ok := readNode(ts)
		if !ok {
			nxt, _ := ts.next()
			return nil, fmt.Errorf("failed to parse token: %v", nxt)
		}
		ns = append(ns, n)
	}

	return ns, nil
}

func readNode(ts *tokenStream) (node, bool) {
	if node, ok := readListNode(ts); ok {
		return node, ok
	}

	if node, ok := readIntNode(ts); ok {
		return node, ok
	}

	if node, ok := readStringNode(ts); ok {
		return node, ok
	}

	if node, ok := readKeyword(ts); ok {
		return node, ok
	}

	if node, ok := readSymbol(ts); ok {
		return node, ok
	}

	return nil, false
}

func readListNode(ts *tokenStream) (node, bool) {
	idx := ts.position()

	if !ts.nextTokenIs(tokenTypePunctuation, "(") {
		return nil, false
	}
	ts.next()

	nodes := []node{}
	for {
		switch {
		case ts.eof():
			ts.seek(idx)
			return nil, false

		case ts.nextTokenIs(tokenTypePunctuation, ")"):
			ts.next()
			goto done

		default:
			node, ok := readNode(ts)
			if !ok {
				return nil, false
			}
			nodes = append(nodes, node)
		}
	}

done:
	return listNode{nodes}, true
}

func readStringNode(ts *tokenStream) (node, bool) {
	if !ts.nextTokenTypeIs(tokenTypeLiteralString) {
		return nil, false
	}

	t, _ := ts.next()
	return stringNode{t.val}, true
}

func readIntNode(ts *tokenStream) (node, bool) {
	if !ts.nextTokenTypeIs(tokenTypeWord) {
		return nil, false
	}

	t, _ := ts.peek()
	n, err := strconv.ParseInt(t.val, 10, 64)
	if err != nil {
		return nil, false
	}
	ts.next()
	return intNode{n}, true
}

func readKeyword(ts *tokenStream) (node, bool) {
	if !ts.nextTokenTypeIs(tokenTypeWord) {
		return nil, false
	}

	t, _ := ts.peek()
	if t.val[0] != ':' {
		return nil, false
	}
	ts.next()
	return keywordNode{t.val}, true
}

func readSymbol(ts *tokenStream) (node, bool) {
	if !ts.nextTokenTypeIs(tokenTypeWord) {
		return nil, false
	}

	t, _ := ts.next()
	return symbolNode{t.val}, true
}
