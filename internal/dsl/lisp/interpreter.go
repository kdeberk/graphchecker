package lisp

import (
	"fmt"

	"dberk.nl/graphchecker/internal/model"
)

// Interpret converts takes an AST and constructs the ioco-model.
// - it interprets the defmessage calls and generates model.Message objects
// - it interprets the defprocess calls and constructs the graphs
func Interpret(as []ast) (*model.Model, error) {
	// TODO: resolve all message references, and state transitions
	return interpretToplevel(as)
}

func interpretToplevel(ns []ast) (*model.Model, error) {
	messages := []*model.Message{}
	processes := []*model.Process{}

	for _, n := range ns {
		switch n := n.(type) {
		case listNode:
			fnCall, err := parseFnCall(n.nodes)
			if err != nil {
				return nil, fmt.Errorf("parseFnCall: %w", err)
			}

			switch fnCall.fnName() {
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

		if fieldCall.fnName() != "field" {
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
func defprocess_body(ns []node, b *processBuilder) error {
	for _, n := range ns {
		if err := defprocess_body_expression(n, b); err != nil {
			return err
		}
	}
	return nil
}

func defprocess_body_expression(n node, b *processBuilder) error {
	switch n := n.(type) {
	case keywordNode:
		state, err := b.allocNamedState(n.name)
		if err != nil {
			return err
		}
		b.addTransition(&model.Transition{From: b.curState, To: state})
		b.curState = state
	case listNode:
		if b.curState == nil {
			return fmt.Errorf("unreachable")
		}

		call, err := parseFnCall(n.nodes)
		if err != nil {
			return err
		}

		switch call.fnName() {
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
			err := defprocess_if(call, b)
			if err != nil {
				return fmt.Errorf("if: %w", err)
			}

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

	if len(valuation) == 0 {
		valuation = nil
	}

	to := b.allocUnnamedState()
	t := &model.Transition{
		From: b.curState,
		To: to,
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
	guard, err := call.nextUnnamedParam().expression()
	if err != nil {
		return err
	}

	then, err := call.nextUnnamedParam().node()
	if err != nil {
		return err
	}

	// TODO: check if params left
	else_, err := call.nextUnnamedParam().node()
	if err != nil {
		// TODO: check if EOF
		// No else defined.
	}

	ifStart := b.curState
	thenStart := b.allocUnnamedState()
	b.addTransition(&model.Transition{
		From: ifStart,
		To: thenStart,
		Send: "then",
		Constraint: guard,
	})

 	b.curState = thenStart
	if err := defprocess_body_expression(then, b); err != nil {
		return err
	}

	thenEnd := b.curState
	b.curState = ifStart

	var elseEnd *model.State;
	if else_ != nil {
		elseStart := b.allocUnnamedState()
		t := &model.Transition{
			From: ifStart,
			To: elseStart,
			Constraint: negateExpression(guard),
		}
		b.addTransition(t)

		b.curState = elseStart
		if err := defprocess_body_expression(else_, b); err != nil {
			return err
		}
		elseEnd = b.curState
	}

	if thenEnd == nil && elseEnd == nil {
		b.curState = nil
		return nil
	}

	ifEnd := b.allocUnnamedState()
	if thenEnd != nil {
		b.addTransition(&model.Transition{
			From: thenEnd,
			To: ifEnd,
			Send: "ifend",
		})
	}

	if elseEnd != nil {
		b.addTransition(&model.Transition{
			From: elseEnd,
			To: ifEnd,
			Send: "elseend",
		})
	}


	b.curState = ifEnd
	return nil
}

