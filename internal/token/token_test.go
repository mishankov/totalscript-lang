package token

import "testing"

func TestLookupIdent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ident    string
		expected TokenType
	}{
		// Keywords
		{"var keyword", "var", VAR},
		{"const keyword", "const", CONST},
		{"function keyword", "function", FUNCTION},
		{"model keyword", "model", MODEL},
		{"enum keyword", "enum", ENUM},
		{"if keyword", "if", IF},
		{"else keyword", "else", ELSE},
		{"switch keyword", "switch", SWITCH},
		{"case keyword", "case", CASE},
		{"default keyword", "default", DEFAULT},
		{"for keyword", "for", FOR},
		{"while keyword", "while", WHILE},
		{"in keyword", "in", IN},
		{"return keyword", "return", RETURN},
		{"break keyword", "break", BREAK},
		{"continue keyword", "continue", CONTINUE},
		{"import keyword", "import", IMPORT},
		{"as keyword", "as", AS},
		{"this keyword", "this", THIS},
		{"is keyword", "is", IS},
		{"true keyword", "true", TRUE},
		{"false keyword", "false", FALSE},
		{"null keyword", "null", NULL},
		{"constructor keyword", "constructor", CONSTRUCTOR},

		// Identifiers (not keywords)
		{"simple identifier", "foo", IDENT},
		{"camelCase identifier", "myVar", IDENT},
		{"snake_case identifier", "my_var", IDENT},
		{"identifier with numbers", "var1", IDENT},
		{"uppercase identifier", "VAR", IDENT},
		{"mixed case identifier", "VaR", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := LookupIdent(tt.ident)
			if result != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, result, tt.expected)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	t.Parallel()
	tok := Token{
		Type:    INTEGER,
		Literal: "42",
		Line:    1,
		Column:  5,
	}

	if tok.Type != INTEGER {
		t.Errorf("tok.Type = %v, want %v", tok.Type, INTEGER)
	}
	if tok.Literal != "42" {
		t.Errorf("tok.Literal = %v, want %v", tok.Literal, "42")
	}
	if tok.Line != 1 {
		t.Errorf("tok.Line = %v, want %v", tok.Line, 1)
	}
	if tok.Column != 5 {
		t.Errorf("tok.Column = %v, want %v", tok.Column, 5)
	}
}
