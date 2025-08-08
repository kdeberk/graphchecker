package lisp

type tokenType string

const (
	tokenTypePunctuation   tokenType = "punctuation"
	tokenTypeLiteralString           = "literalString"
	tokenTypeWord                    = "word"
)

type ast node

type token struct {
	typ   tokenType
	index int
	val   string
}

type node interface {
	Kind() string
}

type listNode struct {
	nodes []node
}

func (_ listNode) Kind() string {
	return "list"
}

var _ node = (*listNode)(nil)

type stringNode struct {
	str string
}

func (_ stringNode) Kind() string {
	return "string"
}

var _ node = (*stringNode)(nil)

type intNode struct {
	int int64
}

func (_ intNode) Kind() string {
	return "int"
}

var _ node = (*intNode)(nil)

type keywordNode struct {
	name string
}

func (_ keywordNode) Kind() string {
	return "keyword"
}

var _ node = (*keywordNode)(nil)

type symbolNode struct {
	name string
}

func (_ symbolNode) Kind() string {
	return "symbol"
}

var _ node = (*symbolNode)(nil)
