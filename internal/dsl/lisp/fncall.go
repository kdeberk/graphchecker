package lisp

import (
	"fmt"

	"dberk.nl/graphchecker/internal/model"
)

// fnCall represents a function call. That is: a list with a symbol as first parameter.
type fnCall struct {
	fName string
	args  []node
	aIdx  int
}

func parseFnCall(ns []node) (*fnCall, error) {
	if len(ns) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	var err error
	call := &fnCall{fName: "", args: ns, aIdx: 0}
	call.fName, err = call.nextUnnamedParam().symbol()

	return call, err
}

func (c *fnCall) fnName() string {
	return c.fName
}

func (c *fnCall) nextParam(name string) *param {
	if c.aIdx == len(c.args) {
		return &param{err: fmt.Errorf("missing required parameter(s)")}
	}

	key, isKey := c.args[c.aIdx].(keywordNode)
	switch {
	case isKey && key.name != name:
		return &param{err: fmt.Errorf("wrong arg name, got %s but expected %s", key.name, name)}

	case isKey && (c.aIdx+1) == len(c.args):
		return &param{err: fmt.Errorf("missing value for arg %s", key.name)}

	case isKey:
		c.aIdx += 1
		fallthrough

	default:
		c.aIdx += 1
		return &param{n: c.args[c.aIdx-1]}
	}
}

func (c *fnCall) nextUnnamedParam() *param {
	if c.aIdx == len(c.args) {
		return &param{err: fmt.Errorf("EOF")}
	}

	c.aIdx += 1
	return &param{n: c.args[c.aIdx-1]}
}

func (c *fnCall) isDone() bool {
	return c.aIdx == len(c.args)
}

type param struct {
	n node
	err  error
}

func (p *param) node() (node, error) {
	if p.err != nil {
		return nil, p.err
	}

	return p.n, nil
}

func (p *param) string() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	strNode, ok := p.n.(stringNode)
	if !ok {
		return "", fmt.Errorf("expected stringNode, got %s", p.n.Kind())
	}

	return strNode.str, nil
}

func (p *param) symbol() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	symNode, ok := p.n.(symbolNode)
	if !ok {
		return "", fmt.Errorf("expected symbolNode, got %s", p.n.Kind())
	}

	return symNode.name, nil
}

func (p *param) keyword() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	keyNode, ok := p.n.(keywordNode)
	if !ok {
		return "", fmt.Errorf("expected keywordNode, got %s", p.n.Kind())
	}

	return keyNode.name, nil
}

func (p *param) list() ([]node, error) {
	if p.err != nil {
		return nil, p.err
	}

	listNode, ok := p.n.(listNode)
	if !ok {
		return nil, fmt.Errorf("expected listNode, got %s", p.n.Kind())
	}

	return listNode.nodes, nil
}

func (p *param) call() (*fnCall, error) {
	ns, err := p.list()
	if err != nil {
		return nil, err
	}

	return parseFnCall(ns)
}

func (p *param) expression() (*model.Expression, error) {
	if p.err != nil {
		return nil, p.err
	}

	return parseExpression(p.n)
}

func (p *param) error() error {
	return p.err
}
