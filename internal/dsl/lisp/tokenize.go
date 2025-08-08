package lisp

import (
	"fmt"
	"unicode"
)

func Tokenize(s string) ([]token, error) {
	tokens := []token{}

	for idx := 0; idx < len(s); {
		r := rune(s[idx])

		if unicode.IsSpace(r) {
			idx++
			continue
		}

		switch {
		case isPunctuation(r):
			tokens = append(tokens, token{tokenTypePunctuation, idx, s[idx : idx+1]})
			idx++

		case r == '"':
			// LiteralString
			jdx := idx + 1
			for ; jdx < len(s); jdx++ {
				if s[jdx] == '"' {
					jdx += 1
					break
				}
			}

			tokens = append(tokens, token{tokenTypeLiteralString, idx, s[idx:jdx]})
			idx = jdx

		case unicode.IsPrint(r):
			// Word
			jdx := idx
			for ; jdx < len(s); jdx++ {
				r_ := rune(s[jdx])
				if unicode.IsSpace(r_) || isPunctuation(r_) {
					break
				}
			}

			tokens = append(tokens, token{tokenTypeWord, idx, s[idx:jdx]})
			idx = jdx

		default:
			return nil, fmt.Errorf("Unknown rune %x", r)
		}
	}

	return tokens, nil
}

func isPunctuation(r rune) bool {
	return r == '(' || r == ')' || r == '{' || r == '}'
}
