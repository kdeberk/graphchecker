package lisp

import (
	"fmt"

	"dberk.nl/graphchecker/internal/model"
)

func Interpret(ns []node) (*model.Model, error) {
	return interpretToplevel(ns)
}

func interpretToplevel(ns []node) (*model.Model, error) {
	messages := []*model.Message{}
	processes := []*model.Process{}

	for _, n := range ns {
		switch n := n.(type) {
		case listNode:
			fnCall, err := parseFnCall(n.nodes)
			if err != nil {
				return nil, fmt.Errorf("parseFnCall: %w", err)
			}

			switch fnCall.fName {
			case "defmessage":
				mess, err := defmessage(fnCall)
				if err != nil {
					return nil, fmt.Errorf("defmessage: %w", err)
				}
				messages = append(messages, mess)

			case "defprocess":
				proc, err := defprocess(fnCall)
				if err != nil {
					return nil, fmt.Errorf("defprocess: %w", err)
				}
				processes = append(processes, proc)

			default:
				return nil, fmt.Errorf("unknown fn: %s", fnCall.fName)
			}
		}
	}

	return &model.Model{Messages: messages, Processes: processes}, nil
}

func defmessage(defCall *fnCall) (*model.Message, error) {
	name, err := defCall.nextParam(":name").symbol()
	if err != nil {
		return nil, err
	}

	fieldNames := []string{}
	for !defCall.isDone() {
		var fieldCall *fnCall

		if len(fieldNames) == 0 {
			fieldCall, err = defCall.nextParam(":fields").call()
		} else {
			fieldCall, err = defCall.nextUnnamedParam().call()
		}

		if err != nil {
			return nil, fmt.Errorf("getting field parameter: %w", err)
		}

		if fieldCall.fName != "field" {
			return nil, fmt.Errorf("expected 'field', got: %s", fieldCall.fName)
		}

		fieldName, err := fieldCall.nextParam(":name").symbol()
		if err != nil {
			return nil, err
		}

		fieldNames = append(fieldNames, fieldName)
	}

	return &model.Message{Name: name, Fields: fieldNames}, nil
}

func defprocess(call *fnCall) (*model.Process, error) {
	// b := newProcessBuilder()

	return nil, nil
}

// TODO: body is a list of nodes, evaluate each one by one
// Can be function or type
func defprocess_body(n node, b *processBuilder) error {
	switch n := n.(type) {
	case keywordNode:
		state, err := b.allocNamedState(n.name)
		if err != nil {
			return err
		}
		b.addTransition(&model.Transition{From: b.curState, To: state})
		b.curState = state
	case listNode:
		call, err := parseFnCall(n.nodes)
		if err != nil {
			return err
		}

		switch call.fName {
		case "!send":
			err := defprocess_send(call, b)
			if err != nil {
				return fmt.Errorf("!send: %w", err)
			}

		case "?receive":
			err := defprocess_receive(call, b)
			if err != nil {
				return fmt.Errorf("?receive: %w", err)
			}

		case "let":
			// transition that initializes variables
			// recurse on contains
			// close scope

		case "if":
			// transition with constraint
			// if 3 parameter, then add transition with inverted scope

		case "select":

		case "goto":
			err := defprocess_goto(call, b)
			if err != nil {
				return fmt.Errorf("goto: %w", err)
			}

		default:
			return fmt.Errorf("unknown fn call %s", call.fName)
		}

	default:
		return fmt.Errorf("unrecognized expression at body")
	}

	return nil
}

func defprocess_send(call *fnCall, b *processBuilder) error {
	if b.curState == nil {
		return fmt.Errorf("unreachable")
	}

	mess, err := call.nextParam(":message").symbol()
	if err != nil {
		return err
	}

	valuation := map[string]*model.Expression{}
	for {
		if call.isDone() {
			break
		}

		name, err := call.nextUnnamedParam().keyword()
		if err != nil {
			return err
		}

		expr, err := call.nextUnnamedParam().expression()
		if err != nil {
			return err
		}

		valuation[name] = expr
	}

	to := b.allocUnnamedState()
	t := &model.Transition{
		Send: mess,
		Valuation: valuation,
	}
	b.addTransition(t)
	b.curState = to
	return nil
}

func defprocess_nameCurrentState(name string, b *processBuilder) error {
	if name == b.initState.Name {
		if b.curState == b.initState {
			// start state was named explicitly, this is fine
			return nil
		}

		return fmt.Errorf("name collision, %s is already taken", name)
	}

	s, err := b.allocNamedState(name)
	if err != nil {
		return err
	}

	if b.curState != nil {
		// If we're at at unreachable state, then this named state is probably
		//  the target of a goto.
		b.addTransition(&model.Transition{
			From: b.curState,
			To: s,
		})
	}

	b.curState = s
	return nil
}

func defprocess_receive(call *fnCall, b *processBuilder) error {
	if b.curState == nil {
		return fmt.Errorf("unreachable")
	}

	mess, err := call.nextParam(":message").symbol()
	if err != nil {
		return err
	}

	to := b.allocUnnamedState()
	t := &model.Transition{
		From: b.curState,
		To: to,
		Receive: mess,
	}
	b.addTransition(t)
	b.curState = to
	return nil
}

func defprocess_goto(call *fnCall, b *processBuilder) error {
	if b.curState == nil {
		return fmt.Errorf("unreachable")
	}

	name, err := call.nextUnnamedParam().keyword()
	if err != nil {
		return err
	}

	to, ok := b.stateForName(name)
	if !ok {
		to, err = b.allocNamedState(name)
		if err != nil {
			return err
		}
	}

	t := &model.Transition{
		From: b.curState,
		To: to,
	}
	b.addTransition(t)
	b.curState = nil
	return nil
}

func defprocess_if(call *fnCall, b *processBuilder) error {
	if b.curState == nil {
		return fmt.Errorf("unreachable")
	}

	// get list as first parameter
	// convert this list to Transition with expression

	// get list as second parameter
	//
	// optionally get list as last parameter
	// convert first list to Transition with negation of expression

	return nil
}

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
		return &param{node: c.args[c.aIdx-1]}
	}
}

func (c *fnCall) nextUnnamedParam() *param {
	if c.aIdx == len(c.args) {
		return &param{err: fmt.Errorf("EOF")}
	}

	c.aIdx += 1
	return &param{node: c.args[c.aIdx-1]}
}

func (c *fnCall) isDone() bool {
	return c.aIdx == len(c.args)
}

type param struct {
	node node
	err  error
}

func (p *param) string() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	strNode, ok := p.node.(stringNode)
	if !ok {
		return "", fmt.Errorf("expected stringNode, got %s", p.node.Kind())
	}

	return strNode.str, nil
}

func (p *param) symbol() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	symNode, ok := p.node.(symbolNode)
	if !ok {
		return "", fmt.Errorf("expected symbolNode, got %s", p.node.Kind())
	}

	return symNode.name, nil
}

func (p *param) keyword() (string, error) {
	if p.err != nil {
		return "", p.err
	}

	keyNode, ok := p.node.(keywordNode)
	if !ok {
		return "", fmt.Errorf("expected keywordNode, got %s", p.node.Kind())
	}

	return keyNode.name, nil
}

func (p *param) list() ([]node, error) {
	if p.err != nil {
		return nil, p.err
	}

	listNode, ok := p.node.(listNode)
	if !ok {
		return nil, fmt.Errorf("expected listNode, got %s", p.node.Kind())
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

	return parseExpression(p.node)
}

func (p *param) error() error {
	return p.err
}

func parseExpression(n node) (*model.Expression, error) {
	switch n := n.(type) {
	case listNode:
		exprs := []*model.Expression{}
		for _, cn := range n.nodes {
			expr, err := parseExpression(cn)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, expr)
		}

		return &model.Expression{
			Type: "lst",
			Sub: exprs,
		}, nil
	case symbolNode:
	  return &model.Expression{
			Type: "ref",
			Ref: n.name,
		}, nil
	case intNode:
		return &model.Expression{
			Type: "int",
			Int: n.int,
		}, nil
	default:
		return nil, fmt.Errorf("unhandled type: %s", n.Kind())
	}
}
