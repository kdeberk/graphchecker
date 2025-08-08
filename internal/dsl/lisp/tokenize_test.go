package lisp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	var tests = []struct {
		str       string
		expTokens []token
		expErr    string
	}{
		{
			str: "()",
			expTokens: []token{
				{tokenTypePunctuation, 0, "("},
				{tokenTypePunctuation, 1, ")"},
			},
		},
		{
			str: "{}",
			expTokens: []token{
				{tokenTypePunctuation, 0, "{"},
				{tokenTypePunctuation, 1, "}"},
			},
		},
		{
			str: "(+ 12 34)",
			expTokens: []token{
				{tokenTypePunctuation, 0, "("},
				{tokenTypeWord, 1, "+"},
				{tokenTypeWord, 3, "12"},
				{tokenTypeWord, 6, "34"},
				{tokenTypePunctuation, 8, ")"},
			},
		},
		{
			str: "(print \"Hello, World!\")",
			expTokens: []token{
				{tokenTypePunctuation, 0, "("},
				{tokenTypeWord, 1, "print"},
				{tokenTypeLiteralString, 7, "\"Hello, World!\""},
				{tokenTypePunctuation, 22, ")"},
			},
		},
		{
			str: "(!send :channel channel :message noResult)",
			expTokens: []token{
				{tokenTypePunctuation, 0, "("},
				{tokenTypeWord, 1, "!send"},
				{tokenTypeWord, 7, ":channel"},
				{tokenTypeWord, 16, "channel"},
				{tokenTypeWord, 24, ":message"},
				{tokenTypeWord, 33, "noResult"},
				{tokenTypePunctuation, 41, ")"},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Tokenize(\"%s\")", test.str), func(t *testing.T) {
			tokens, err := Tokenize(test.str)

			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("Did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("Expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("Expected a different error, got %v", err)

			case !reflect.DeepEqual(tokens, test.expTokens):
				t.Errorf("Output was not expected, got: %v", tokens)
			}
		})
	}
}
