package ast

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/token"
)

func TestString(t *testing.T) {
	t.Parallel()
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Type: nil,
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if program.String() != "var myVar = anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}

func TestVarStatement(t *testing.T) {
	t.Parallel()
	stmt := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "x"},
			Value: "x",
		},
		Type: &TypeExpression{
			Token: token.Token{Type: token.IDENT, Literal: "integer"},
			Name:  "integer",
		},
		Value: &IntegerLiteral{
			Token: token.Token{Type: token.INTEGER, Literal: "5"},
			Value: 5,
		},
	}

	if stmt.TokenLiteral() != "var" {
		t.Errorf("stmt.TokenLiteral() wrong. got=%q", stmt.TokenLiteral())
	}

	expected := "var x: integer = 5"
	if stmt.String() != expected {
		t.Errorf("stmt.String() wrong. expected=%q, got=%q", expected, stmt.String())
	}
}

func TestConstStatement(t *testing.T) {
	t.Parallel()
	stmt := &ConstStatement{
		Token: token.Token{Type: token.CONST, Literal: "const"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "PI"},
			Value: "PI",
		},
		Value: &FloatLiteral{
			Token: token.Token{Type: token.FLOAT, Literal: "3.14"},
			Value: 3.14,
		},
	}

	if stmt.TokenLiteral() != "const" {
		t.Errorf("stmt.TokenLiteral() wrong. got=%q", stmt.TokenLiteral())
	}

	expected := "const PI = 3.14"
	if stmt.String() != expected {
		t.Errorf("stmt.String() wrong. expected=%q, got=%q", expected, stmt.String())
	}
}

func TestReturnStatement(t *testing.T) {
	t.Parallel()
	stmt := &ReturnStatement{
		Token: token.Token{Type: token.RETURN, Literal: "return"},
		ReturnValue: &IntegerLiteral{
			Token: token.Token{Type: token.INTEGER, Literal: "42"},
			Value: 42,
		},
	}

	if stmt.TokenLiteral() != "return" {
		t.Errorf("stmt.TokenLiteral() wrong. got=%q", stmt.TokenLiteral())
	}

	expected := "return 42"
	if stmt.String() != expected {
		t.Errorf("stmt.String() wrong. expected=%q, got=%q", expected, stmt.String())
	}
}

func TestInfixExpression(t *testing.T) {
	t.Parallel()
	expr := &InfixExpression{
		Token: token.Token{Type: token.PLUS, Literal: "+"},
		Left: &IntegerLiteral{
			Token: token.Token{Type: token.INTEGER, Literal: "5"},
			Value: 5,
		},
		Operator: "+",
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INTEGER, Literal: "10"},
			Value: 10,
		},
	}

	expected := "(5 + 10)"
	if expr.String() != expected {
		t.Errorf("expr.String() wrong. expected=%q, got=%q", expected, expr.String())
	}
}

func TestPrefixExpression(t *testing.T) {
	t.Parallel()
	expr := &PrefixExpression{
		Token:    token.Token{Type: token.MINUS, Literal: "-"},
		Operator: "-",
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INTEGER, Literal: "5"},
			Value: 5,
		},
	}

	expected := "(-5)"
	if expr.String() != expected {
		t.Errorf("expr.String() wrong. expected=%q, got=%q", expected, expr.String())
	}
}

func TestArrayLiteral(t *testing.T) {
	t.Parallel()
	arr := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INTEGER, Literal: "1"},
				Value: 1,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INTEGER, Literal: "2"},
				Value: 2,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INTEGER, Literal: "3"},
				Value: 3,
			},
		},
	}

	expected := "[1, 2, 3]"
	if arr.String() != expected {
		t.Errorf("arr.String() wrong. expected=%q, got=%q", expected, arr.String())
	}
}

func TestTypeExpression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeExpr *TypeExpression
		expected string
	}{
		{
			name: "simple type",
			typeExpr: &TypeExpression{
				Name: "integer",
			},
			expected: "integer",
		},
		{
			name: "optional type",
			typeExpr: &TypeExpression{
				Name:     "string",
				Optional: true,
			},
			expected: "string?",
		},
		{
			name: "generic type",
			typeExpr: &TypeExpression{
				Name:    "array",
				Generic: []string{"integer"},
			},
			expected: "array<integer>",
		},
		{
			name: "union type",
			typeExpr: &TypeExpression{
				Union: []string{"integer", "string"},
			},
			expected: "integer | string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.typeExpr.String()
			if result != tt.expected {
				t.Errorf("type.String() wrong. expected=%q, got=%q", tt.expected, result)
			}
		})
	}
}
