package model

import (
	"fmt"
	"strings"
)

type Model struct {
	Messages  []*Message
	Processes []*Process
}

type Message struct {
	Name   string
	Fields []string
}

func (m *Message) String() string {
	if len(m.Fields) == 0 {
		return fmt.Sprintf("(defmessage %s)", m.Name)
	}

	fields := []string{}
	for _, f := range m.Fields {
		fields = append(fields, fmt.Sprintf("(field %s)", f))
	}

	return fmt.Sprintf("(defmessage %s) %s", m.Name, strings.Join(fields, " "))
}

type Process struct {
	Name        string
	Vars        []string
	States      []string
	Transitions []*Transition
}

type State struct {
	ID   int
	Name string
}

func (s *State) Named() bool {
	return s.Name != ""
}

type Transition struct {
	From, To *State
	Receive string
	Send string
	Valuation map[string]*Expression
	Constraint *Expression
}

type Variable struct {
	ID   int
	Name string
}

type Expression struct {
	Type string
	Ref string
	Sub []*Expression
	Int int64
}
