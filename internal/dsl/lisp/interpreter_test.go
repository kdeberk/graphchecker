package lisp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"dberk.nl/graphchecker/internal/model"
	"github.com/stretchr/testify/assert"
)



func TestDefmessage(t *testing.T) {
	var tests = []struct {
		name string
		str string
		expMessage *model.Message
		expErr string
	}{
		{
			name: "implicit name, no fields",
			str: "(defmessage MessageName)",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{},
			},
		},
		{
			name: "explicit name, no fields",
			str: "(defmessage :name MessageName)",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{},
			},
		},
		{
			name: "implicit name, single field",
			str: "(defmessage MessageName (field FieldOne))",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{"FieldOne"},
			},
		},
		{
			name: "implicit name, single field with explicit name",
			str: "(defmessage MessageName (field :name FieldOne))",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{"FieldOne"},
			},
		},
		{
			name: "implicit name, multiple fields",
			str: "(defmessage MessageName (field FieldOne) (field FieldTwo))",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{"FieldOne", "FieldTwo"},
			},
		},
		{
			name: "explicit name, explicit multiple fields",
			str: "(defmessage :name MessageName :fields (field FieldOne) (field FieldTwo))",
			expMessage: &model.Message{
				Name: "MessageName", Fields: []string{"FieldOne", "FieldTwo"},
			},
		},
		{
			name: "missing explicit name",
			str: "(defmessage :name)",
			expErr: "missing value for arg :name",
		},
		{
			name: "unknown field",
			str: "(defmessage :foo)",
			expErr: "wrong arg name, got :foo but expected :name",
		},
		{
			name: "no parameters",
			str: "(defmessage)",
			expErr: "missing required parameter(s)",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("defmessage - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			mess, err := defmessage(call)

			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(mess, test.expMessage):
				t.Errorf("output was not expected, got: %v", mess)
			}
		})
	}
}

func TestSend(t *testing.T) {
	var tests = []struct {
		name string
		str string
		inProcessBuilder func () *processBuilder
		expProcessBuilder func () *processBuilder
		expTransition *model.Transition
		expErr string
	}{
		{
			name: "message implicit name, no parameters",
			str: "(!send MessageName)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				to := b.allocUnnamedState()
				b.addTransition(&model.Transition{
					From: b.initState,
					To: to,
					Send: "MessageName",
					Valuation: map[string]*model.Expression{},
				})
				b.curState = to
				return b
			},
			// expTransition: &model.Transition{
			// 	Send: "MessageName",
			// 	Valuation: map[string]*model.Expression{},
			// },
		},
		{
			name: "message explicit name, no parameters",
			str: "(!send :message MessageName)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				to := b.allocUnnamedState()
				b.addTransition(&model.Transition{
					From: b.initState,
					To: to,
					Send: "MessageName",
					Valuation: map[string]*model.Expression{},
				})
				b.curState = to
				return b
			},

			// expTransition: &model.Transition{
			// 	Send: "MessageName",
			// 	Valuation: map[string]*model.Expression{},
			// },
		},
		{
			name: "message explicit name, with parameters",
			str: "(!send :message MessageName :fieldOne (+ 1 var-one) :fieldTwo var-two)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				to := b.allocUnnamedState()
				b.addTransition(&model.Transition{
					From: b.initState,
					To: to,
					Send: "MessageName",
					Valuation: map[string]*model.Expression{
						":fieldOne": {
							Type: "lst",
							Sub: []*model.Expression{
								{
									Type: "ref",
									Ref: "+",
								},
								{
									Type: "int",
									Int: 1,
								},
								{
									Type: "ref",
									Ref: "var-one",
								},
							},
						},
						":fieldTwo": {
							Type: "ref",
							Ref: "var-two",
						},
					},
				})
				b.curState = to
				return b
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("!send - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			b := test.inProcessBuilder()
			err = defprocess_send(call, b)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(b, test.expProcessBuilder()):
				t.Errorf("output was not expected; expected %v, got: %v", test.expProcessBuilder(), b)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	var tests = []struct {
		name string
		str string
		inProcessBuilder func () *processBuilder
		expProcessBuilder func () *processBuilder
		expErr string
	}{
		{
			name: "message implicit name, no update",
			str: "(?receive MessageName)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				to := b.allocUnnamedState()
				b.addTransition(&model.Transition{
					From: b.initState,
					To: to,
					Receive: "MessageName",
				})
				b.curState = to
				return b
			},
		},
		{
			name: "message explicit name, no parameters",
			str: "(?receive :message MessageName)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				to := b.allocUnnamedState()
				b.addTransition(&model.Transition{
					From: b.initState,
					To: to,
					Receive: "MessageName",
				})
				b.curState = to
				return b
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("?receive - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			b := test.inProcessBuilder()
			err = defprocess_receive(call, b)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(b, test.expProcessBuilder()):
				t.Errorf("output was not expected; expected %v, got: %v", test.expProcessBuilder(), b)
			}
		})
	}
}

func TestGoto(t *testing.T) {
	var tests = []struct {
		name string
		str string
		inProcessBuilder func () *processBuilder
		expProcessBuilder func () *processBuilder
		expErr string
	}{
		{
			name: "unknown state",
			str: "(goto :unknown-state)",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				b.allocNamedState(":unknown-state")
				b.addTransition(&model.Transition{
					From: b.curState, // init state
					To: b.namedStates[":unknown-state"],
				})
				b.curState = nil
				return b
			},
		},
		{
			name: "known state",
			str: "(goto :known-state)",
			inProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				b.allocNamedState(":known-state")
				return b
			},
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				b.allocNamedState(":known-state")
				b.addTransition(&model.Transition{
					From: b.curState, // init state
					To: b.namedStates[":known-state"],
				})
				b.curState = nil
				return b
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("goto - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			b := test.inProcessBuilder()
			err = defprocess_goto(call, b)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(b, test.expProcessBuilder()):
				t.Errorf("output was not expected, got: %v", b)
			}
		})
	}
}

func TestNameCurrentState(t *testing.T) {
	var tests = []struct {
		name string
		str string
		inProcessBuilder func () *processBuilder
		expProcessBuilder func () *processBuilder
		expErr string
	}{
		{
			name: "from unreachable state",
			str: ":some-state",
			inProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				b.curState = nil
				return b
			},
			expProcessBuilder: func () *processBuilder {
				b := newProcessBuilder()
				st, _ := b.allocNamedState(":some-state")
				b.curState = st
				return b
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("nameCurrentState - %s", test.name), func(t *testing.T) {
			b := test.inProcessBuilder()
			err := defprocess_nameCurrentState(test.str, b)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(b, test.expProcessBuilder()):
				t.Errorf("output was not expected, got: %v", b)
			}
		})
	}
}

func TestIf(t *testing.T) {
	var tests = []struct {
		name string
		str string
		inProcessBuilder func () *processBuilder
		expProcessBuilder func () *processBuilder
		expErr string
	}{
		{
			name: "if statement",
			str: "(if (= foo 1) (!send :message MessageA) (!send :message MessageB))",
			inProcessBuilder: newProcessBuilder,
			expProcessBuilder: func() *processBuilder {
				b := newProcessBuilder()
				ifStart := b.curState
				thenStart := b.allocUnnamedState()
				thenEnd := b.allocUnnamedState()
				elseStart := b.allocUnnamedState()
				elseEnd := b.allocUnnamedState()
				ifEnd := b.allocUnnamedState()

				b.addTransition(&model.Transition{
					From: ifStart,
					To: thenStart,
					Send: "then",
					Constraint: &model.Expression{
						Type: "lst",
						Sub: []*model.Expression{
							{Type: "ref", Ref: "="},
							{Type: "ref", Ref: "foo"},
							{Type: "int", Int: 1},
						},
					},
				})
				b.addTransition(&model.Transition{
					From: thenStart,
					To: thenEnd,
					Send: "MessageA",
				})

				b.addTransition(&model.Transition{
					From: ifStart,
					To: elseStart,
					Constraint: &model.Expression{
						Type: "lst",
						Sub: []*model.Expression{
							{Type: "ref", Ref: "not"},
							{Type: "lst",
								Sub: []*model.Expression{
									{Type: "ref", Ref: "="},
									{Type: "ref", Ref: "foo"},
									{Type: "int", Int: 1},
								},
							},
						},
					},
				})
				b.addTransition(&model.Transition{
					From: elseStart,
					To: elseEnd,
					Send: "MessageB",
				})

				b.addTransition(&model.Transition{
					From: thenEnd,
					To: ifEnd,
					Send: "ifend",
				})
				b.addTransition(&model.Transition{
					From: elseEnd,
					To: ifEnd,
					Send: "elseend",
				})
				b.curState = ifEnd
				return b
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("if - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			b := test.inProcessBuilder()
			err = defprocess_if(call, b)

			if test.expErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expErr)
			} else {
				assert.Equal(t, nil, err, "expected err to be nil, got %v")
				assert.Equal(t, test.expProcessBuilder(), b, "expected builder %v to be equal to %v")
			}
		})
	}
}

func asFnCall(s string) (*fnCall, error) {
	tokens, err := Tokenize(s)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %v", err)
	}

	nodes, err := ParseTokenStream(tokens)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %v", err)
	}

	if(len(nodes) == 0) {
		return nil, fmt.Errorf("empty list")
	} else if(1 < len(nodes)) {
		return nil, fmt.Errorf("expected single item")
	}

	listNode, ok := nodes[0].(listNode)
	if !ok {
		return nil, fmt.Errorf("not a listNode")
	}

	return parseFnCall(listNode.nodes)
}
