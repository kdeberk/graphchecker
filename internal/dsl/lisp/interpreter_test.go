package lisp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"dberk.nl/graphchecker/internal/model"
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
		expTransition *model.Transition
		expErr string
	}{
		{
			name: "message implicit name, no parameters",
			str: "(!send MessageName)",
			expTransition: &model.Transition{
				Message: "MessageName",
				Valuation: map[string]*model.Expression{},
			},
		},
		{
			name: "message explicit name, no parameters",
			str: "(!send :message MessageName)",
			expTransition: &model.Transition{
				Message: "MessageName",
				Valuation: map[string]*model.Expression{},
			},
		},
		{
			name: "message explicit name, with parameters",
			str: "(!send :message MessageName :fieldOne (+ 1 var-one) :fieldTwo var-two)",
			expTransition: &model.Transition{
				Message: "MessageName",
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
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("!send - %s", test.name), func(t *testing.T) {
			call, err := asFnCall(test.str)
			if err != nil {
				t.Errorf("didn't expect to fail: %v", err)
			}

			tr, err := defprocess_send(call)
			switch {
			case test.expErr == "" && err != nil:
				t.Errorf("did not expect failure, got: %v", err)

			case test.expErr != "" && err == nil:
				t.Errorf("expected error, did not get any")

			case test.expErr != "" && !strings.Contains(err.Error(), test.expErr):
				t.Errorf("expected a different error, got %v", err)

			case test.expErr == "" && !reflect.DeepEqual(tr, test.expTransition):
				t.Errorf("output was not expected, got: %v", tr)
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
