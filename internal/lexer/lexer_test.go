package lexer

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/token"
)

func TestNextToken_SimpleTokens(t *testing.T) {
	input := `=+-*/%(){}[],;:.`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.COLON, ":"},
		{token.DOT, "."},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_TwoCharTokens(t *testing.T) {
	input := `== != <= >= && || .. ** // += -= *= /= %=`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.EQ, "=="},
		{token.NOT_EQ, "!="},
		{token.LT_EQ, "<="},
		{token.GT_EQ, ">="},
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.DOTDOT, ".."},
		{token.POWER, "**"},
		{token.SLASHSLASH, "//"},
		{token.PLUS_ASSIGN, "+="},
		{token.MINUS_ASSIGN, "-="},
		{token.ASTERISK_ASSIGN, "*="},
		{token.SLASH_ASSIGN, "/="},
		{token.PERCENT_ASSIGN, "%="},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_ThreeCharTokens(t *testing.T) {
	input := `..=`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DOTDOTEQ, "..="},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Numbers(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{"integer", "42", token.INTEGER, "42"},
		{"zero", "0", token.INTEGER, "0"},
		{"large integer", "123456789", token.INTEGER, "123456789"},
		{"float", "3.14", token.FLOAT, "3.14"},
		{"float with zero", "0.5", token.FLOAT, "0.5"},
		{"float with trailing zeros", "1.00", token.FLOAT, "1.00"},
		{"integer not float", "42.", token.INTEGER, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Errorf("tokentype wrong. expected=%q, got=%q", tt.expectedType, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestNextToken_Strings(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{"simple string", `"hello"`, "hello"},
		{"empty string", `""`, ""},
		{"string with spaces", `"hello world"`, "hello world"},
		{"string with newline escape", `"hello\nworld"`, "hello\nworld"},
		{"string with tab escape", `"hello\tworld"`, "hello\tworld"},
		{"string with quote escape", `"say \"hello\""`, `say "hello"`},
		{"string with backslash escape", `"path\\to\\file"`, `path\to\file`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != token.STRING {
				t.Errorf("tokentype wrong. expected=%q, got=%q", token.STRING, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestNextToken_IdentifiersAndKeywords(t *testing.T) {
	input := `var const function model enum if else switch case default
	for while in return break continue import as this is true false null`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.CONST, "const"},
		{token.FUNCTION, "function"},
		{token.MODEL, "model"},
		{token.ENUM, "enum"},
		{token.IF, "if"},
		{token.ELSE, "else"},
		{token.SWITCH, "switch"},
		{token.CASE, "case"},
		{token.DEFAULT, "default"},
		{token.FOR, "for"},
		{token.WHILE, "while"},
		{token.IN, "in"},
		{token.RETURN, "return"},
		{token.BREAK, "break"},
		{token.CONTINUE, "continue"},
		{token.IMPORT, "import"},
		{token.AS, "as"},
		{token.THIS, "this"},
		{token.IS, "is"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.NULL, "null"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Identifiers(t *testing.T) {
	input := `foo bar myVar my_var var1 _private`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.IDENT, "bar"},
		{token.IDENT, "myVar"},
		{token.IDENT, "my_var"},
		{token.IDENT, "var1"},
		{token.IDENT, "_private"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_SingleLineComments(t *testing.T) {
	input := `# This is a comment
var x = 5 # inline comment
# another comment`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INTEGER, "5"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_MultiLineComments(t *testing.T) {
	input := `###
This is a
multi-line comment
###
var x = 5`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INTEGER, "5"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LineAndColumn(t *testing.T) {
	input := `var x = 5
var y = 10`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{
		{token.VAR, "var", 1, 1},
		{token.IDENT, "x", 1, 5},
		{token.ASSIGN, "=", 1, 7},
		{token.INTEGER, "5", 1, 9},
		{token.VAR, "var", 2, 1},
		{token.IDENT, "y", 2, 5},
		{token.ASSIGN, "=", 2, 7},
		{token.INTEGER, "10", 2, 9},
		{token.EOF, "", 2, 11},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}

		if tok.Line != tt.expectedLine {
			t.Errorf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Line)
		}

		if tok.Column != tt.expectedColumn {
			t.Errorf("tests[%d] - column wrong. expected=%d, got=%d",
				i, tt.expectedColumn, tok.Column)
		}
	}
}

func TestNextToken_CompleteProgram(t *testing.T) {
	input := `var a = 3
const add = function (x, y) {
  return x + y
}

var result = add(5, 10)`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INTEGER, "3"},
		{token.CONST, "const"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "function"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.RBRACE, "}"},
		{token.VAR, "var"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.INTEGER, "5"},
		{token.COMMA, ","},
		{token.INTEGER, "10"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_SpecialCharacters(t *testing.T) {
	input := `? | @ !`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.QUESTION, "?"},
		{token.PIPE, "|"},
		{token.AT, "@"},
		{token.NOT, "!"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
