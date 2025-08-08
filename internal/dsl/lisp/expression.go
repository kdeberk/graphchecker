package lisp

import (
	"fmt"

	"dberk.nl/graphchecker/internal/model"
)

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

func negateExpression(expr *model.Expression) *model.Expression {
	exprs := []*model.Expression{
		{
			Type: "ref",
			Ref: "not",
		},
		expr,
	}

	return &model.Expression{
		Type: "lst",
		Sub: exprs,
	}
}
