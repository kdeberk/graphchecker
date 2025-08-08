package lisp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseTokenStream(t *testing.T) {
	var tests = []struct {
		str     string
		expNode []node
		expErr  string
	}{
		{
			str:     "()",
			expNode: []node{listNode{[]node{}}},
		},
		{
			str: "(+ 12 34)",
			expNode: []node{
				listNode{
					[]node{
						symbolNode{"+"},
						intNode{12},
						intNode{34},
					}},
			},
		},
		{
			str: "(print \"Hello, World!\")",
			expNode: []node{
				listNode{
					[]node{
						symbolNode{"print"},
						stringNode{"\"Hello, World!\""},
					},
				},
			},
		},
		{
			str: "(!send :channel channel :message noResult)",
			expNode: []node{
				listNode{
					[]node{
						symbolNode{"!send"},
						keywordNode{":channel"},
						symbolNode{"channel"},
						keywordNode{":message"},
						symbolNode{"noResult"},
					}},
			},
		},
		{
			str: "(defun ! (n) (if (<= n 1) 1 (! (- n 1))))",
			expNode: []node{
				listNode{
					[]node{
						symbolNode{"defun"},
						symbolNode{"!"},
						listNode{
							[]node{
								symbolNode{"n"},
							},
						},
						listNode{
							[]node{
								symbolNode{"if"},
								listNode{
									[]node{
										symbolNode{"<="},
										symbolNode{"n"},
										intNode{1},
									},
								},
								intNode{1},
								listNode{
									[]node{
										symbolNode{"!"},
										listNode{
											[]node{
												symbolNode{"-"},
												symbolNode{"n"},
												intNode{1},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ParseTokenStream(\"%s\")", test.str), func(t *testing.T) {
			tokens, err := Tokenize(test.str)
			if err != nil {
				t.Errorf("did not expect tokenization to fail: %v", err)
			}

			node, err := ParseTokenStream(tokens)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("Did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("Expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("Expected a different error, got %v", err)

			case !reflect.DeepEqual(node, test.expNode):
				t.Errorf("Output was not expected, got: %v", node)
			}
		})
	}
}
