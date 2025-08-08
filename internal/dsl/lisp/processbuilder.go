package lisp

import (
	"fmt"

	"dberk.nl/graphchecker/internal/model"
)

type processBuilder struct {
	stateCounter, variableCounter int
	initState *model.State
	curState                      *model.State
	states                        []*model.State
	transitions                   []*model.Transition
	namedStates                   map[string]*model.State
	variables                     map[string]*model.Variable
	scopes                        []map[string]*model.Variable
}

func newProcessBuilder() *processBuilder {
	p := &processBuilder{
		stateCounter: 1,
		initState:    nil,
		curState:     nil,
		states:       []*model.State{},
		transitions:  []*model.Transition{},
		namedStates:  map[string]*model.State{},
	}


	p.initState, _ = p.allocNamedState(":start");
	p.curState = p.initState
	return p
}

func (b *processBuilder) allocUnnamedState() *model.State {
	s := &model.State{ID: b.stateCounter, Name: ""}
	b.states = append(b.states, s)
	b.stateCounter++
	return s
}

func (b *processBuilder) allocNamedState(name string) (*model.State, error) {
	if _, ok := b.namedStates[name]; ok {
		return nil, fmt.Errorf("state already known")
	}

	s := &model.State{ID: b.stateCounter, Name: name}
	b.namedStates[name] = s
	b.states = append(b.states, s)
	b.stateCounter++
	return s, nil
}

func (b *processBuilder) addTransition(t *model.Transition) {
	b.transitions = append(b.transitions, t)
}

func (b *processBuilder) setCurState(s *model.State) {
	b.curState = s
}

func (b *processBuilder) stateForName(n string) (*model.State, bool) {
	state, ok := b.namedStates[n]
	return state, ok
}

func (b *processBuilder) openLexicalScope() {
	b.scopes = append(b.scopes, map[string]*model.Variable{})
}

func (b *processBuilder) closeLexicalScope() {
	b.scopes = b.scopes[:len(b.scopes)-1]
}

func (b *processBuilder) allocVariable(name string) {
	v := &model.Variable{ID: b.variableCounter, Name: name}
	b.variables[name] = v
	b.variableCounter++
	b.scopes[len(b.scopes)-1][name] = v
}

func (b *processBuilder) resolveVariable(name string) (*model.Variable, error) {
	for idx := len(b.scopes) - 1; 0 <= idx; idx-- {
		if v, ok := b.scopes[idx][name]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("could not resolve variable %s", name)
}
